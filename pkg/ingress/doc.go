/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ingress

// This package contains the interface is required to implement to build an Ingress controller
// A dummy implementation could be
//
// func main() {
//  dc := newDummyController()
//	controller.NewIngressController(dc)
//	glog.Infof("shutting down Ingress controller...")
// }
//
//where newDummyController returns and implementation of the Controller interface:
//
// func newDummyController() ingress.Controller {
//   return &DummyController{}
// }
//
// type DummyController struct {
// }
//
// func (dc DummyController) Reload(data []byte) ([]byte, error) {
//   err := ioutil.WriteFile("/arbitrary-path", data, 0644)
//   if err != nil {
//     return nil, err
//   }
//
//   return exec.Command("some command", "--reload").CombinedOutput()
// }
//
// func (dc DummyController) Test(file string) *exec.Cmd {
//     return exec.Command("some command", "--config-file", file)
// }
//
// func (dc DummyController) OnUpdate(*api.ConfigMap, Configuration) ([]byte, error) {
//     return []byte(`<string containing a configuration file>`)
// }
//
// func (dc DummyController) BackendDefaults() defaults.Backend {
//     return ingress.NewStandardDefaults()
// }
//
// func (n DummyController) Name() string {
// 	   return "dummy Controller"
// }
//
// func (n DummyController) Check(_ *http.Request) error {
//     return nil
// }
//
// func (dc DummyController) Info() *BackendInfo {
//     Name: "dummy",
//   Release: "0.0.0",
//   Build: "git-00000000",
//   Repository: "git://foo.bar.com",
// }
