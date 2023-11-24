package controller

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	networking "k8s.io/api/networking/v1"
	"k8s.io/ingress-nginx/internal/ingress/annotations"
	"k8s.io/ingress-nginx/internal/ingress/controller/store"
	"k8s.io/ingress-nginx/pkg/apis/ingress"
	"k8s.io/klog/v2"
)

type Namespace string
type Name string

const (
	// time of batch collecting
	admissionDelaySeconds = 3
	admissionDelay        = admissionDelaySeconds * time.Second

	// amount of concurrent batch consumers
	batchConsumerCount = 30 / admissionDelaySeconds
)

type AdmissionBatcher struct {
	ingresses     []*networking.Ingress
	errorChannels []chan error

	// flag for consumer goroutine indicating whether it should keep processing or not
	isWorking bool

	// mutex protecting queues access
	mu *sync.Mutex

	// wait group to monitor consumer goroutine lifetime
	consumerWG sync.WaitGroup

	// when was last not empty batch consumed by some worker for validation
	lastBatchConsumedTime time.Time
}

func NewAdmissionBatcher() AdmissionBatcher {
	return AdmissionBatcher{
		ingresses:     nil,
		errorChannels: nil,
		isWorking:     true,
		mu:            &sync.Mutex{},
		consumerWG:    sync.WaitGroup{},
	}
}

func (n *NGINXController) StartAdmissionBatcher() {
	for i := 0; i < batchConsumerCount; i++ {
		go n.AdmissionBatcherConsumerRoutine()
	}
}

func (n *NGINXController) StopAdmissionBatcher() {
	n.admissionBatcher.mu.Lock()
	defer n.admissionBatcher.mu.Unlock()

	n.admissionBatcher.isWorking = false
}

// AdmissionBatcherConsumerRoutine is started during ingress-controller startup phase
// And it should stop during ingress-controller's graceful shutdown
func (n *NGINXController) AdmissionBatcherConsumerRoutine() {
	n.admissionBatcher.consumerWG.Add(1)
	defer n.admissionBatcher.consumerWG.Done()

	klog.Info("Admission batcher routine started")

	// prevent races on isWorking field
	n.admissionBatcher.mu.Lock()
	for n.admissionBatcher.isWorking {
		timeSinceLastBatchPull := time.Now().Sub(n.admissionBatcher.lastBatchConsumedTime)
		n.admissionBatcher.mu.Unlock()

		time.Sleep(max(time.Duration(0), admissionDelay-timeSinceLastBatchPull))
		newIngresses, errorChannels := n.admissionBatcher.fetchNewBatch()
		if len(newIngresses) != 0 {
			err := n.validateNewIngresses(newIngresses)
			for _, erCh := range errorChannels {
				erCh <- err
			}
		}

		n.admissionBatcher.mu.Lock()
	}

	klog.Info("Admission batcher routine finished")
}

func groupByNamespacesAndNames(ingresses []*networking.Ingress) map[Namespace]map[Name]struct{} {
	grouped := make(map[Namespace]map[Name]struct{})

	for _, ing := range ingresses {
		ns := Namespace(ing.ObjectMeta.Namespace)
		name := Name(ing.ObjectMeta.Name)
		if _, exists := grouped[ns]; !exists {
			grouped[ns] = make(map[Name]struct{})
		}

		grouped[ns][name] = struct{}{}
	}

	return grouped
}

func (n *NGINXController) validateNewIngresses(newIngresses []*networking.Ingress) error {
	cfg := n.store.GetBackendConfiguration()
	cfg.Resolver = n.resolver

	newIngsDict := groupByNamespacesAndNames(newIngresses)

	allIngresses := n.store.ListIngresses()
	ings := store.FilterIngresses(allIngresses, func(toCheck *ingress.Ingress) bool {
		ns := Namespace(toCheck.ObjectMeta.Namespace)
		name := Name(toCheck.ObjectMeta.Name)

		nsNames, nsMatchesAny := newIngsDict[ns]
		if !nsMatchesAny {
			return false
		}
		_, nameMatchesAny := nsNames[name]
		return nameMatchesAny
	})

	var ingsListSB strings.Builder
	for _, ing := range newIngresses {
		ingsListSB.WriteString(fmt.Sprintf("%v/%v ", ing.Namespace, ing.Name))
	}
	ingsListStr := ingsListSB.String()

	annotationsExtractor := annotations.NewAnnotationExtractor(n.store)
	for _, ing := range newIngresses {
		ann, err := annotationsExtractor.Extract(ing)
		if err != nil {
			return err
		}
		ings = append(ings, &ingress.Ingress{
			Ingress:           *ing,
			ParsedAnnotations: ann,
		})
	}
	//debug
	klog.Info("New ingresses with annotations appended for ", ingsListStr)

	start := time.Now()
	_, _, newIngCfg := n.getConfiguration(ings)
	//debug
	klog.Info("Configuration generated in ", time.Now().Sub(start).Seconds(), " seconds for ", ingsListStr)

	start = time.Now()
	template, err := n.generateTemplate(cfg, *newIngCfg)
	if err != nil {
		return errors.Wrap(err, "error while validating batch of ingresses")
	}
	//debug
	klog.Info("Generated nginx template in ", time.Now().Sub(start).Seconds(), " seconds for ", ingsListStr)

	start = time.Now()
	err = n.testTemplate(template)
	if err != nil {
		return errors.Wrap(err, "error while validating batch of ingresses")
	}
	//debug
	klog.Info("Tested nginx template in ", time.Now().Sub(start).Seconds(), " seconds for ", ingsListStr)

	return nil
}

func (ab *AdmissionBatcher) fetchNewBatch() (ings []*networking.Ingress, errorChannels []chan error) {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	if len(ab.ingresses) == 0 {
		return nil, nil
	}

	ings = ab.ingresses
	errorChannels = ab.errorChannels

	// debug
	var sb strings.Builder
	sb.WriteString("Fetched new batch of ingresses: ")
	for _, ing := range ings {
		sb.WriteString(fmt.Sprintf("%s/%s ", ing.Namespace, ing.Name))
	}
	klog.Info(sb.String())

	ab.errorChannels = nil
	ab.ingresses = nil

	ab.lastBatchConsumedTime = time.Now()

	return ings, errorChannels
}

func (ab *AdmissionBatcher) ValidateIngress(ing *networking.Ingress) error {
	ab.mu.Lock()

	ab.ingresses = append(ab.ingresses, ing)

	errCh := make(chan error)
	ab.errorChannels = append(ab.errorChannels, errCh)

	ab.mu.Unlock()

	// debug
	klog.Info("Ingress ", fmt.Sprintf("%v/%v", ing.Namespace, ing.Name), " submitted for batch validation, waiting for verdict...")
	return <-errCh
}

func max(d1, d2 time.Duration) time.Duration {
	if d1 < d2 {
		return d2
	}
	return d1
}
