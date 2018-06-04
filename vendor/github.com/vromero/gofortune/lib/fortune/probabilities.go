package fortune

import (
	"math/rand"
)

func GetRandomLeafNode(fsDescriptor FileSystemNodeDescriptor) FileSystemNodeDescriptor {
	if len(fsDescriptor.Children) > 0 {
		r := rand.Float32() * fsDescriptor.Percent
		var cumulativeProbability float32

		for i := range fsDescriptor.Children {
			cumulativeProbability += fsDescriptor.Children[i].Percent
			if r <= cumulativeProbability {
				return GetRandomLeafNode(fsDescriptor.Children[i])
			}
		}
		panic("No branch was randomly selected")
	} else {
		return fsDescriptor
	}
}

// SetProbabilities calculates percentage of possibility of being randomly chosen
// for each of the nodes of a FileSystemDescriptor graph.
func SetProbabilities(fsDescriptor *FileSystemNodeDescriptor, considerEqualSize bool) {
	if considerEqualSize {
		setProbabilitiesEqualSize(fsDescriptor, fsDescriptor)
	} else {
		calculateUndefinedProbability(fsDescriptor)
		setProbabilities(fsDescriptor, fsDescriptor)
	}
}

// setProbabilitiesEqualSize calculates percentage of possibility of being randomly chosen
// for each of the nodes of a FileSystemDescriptor graph assuming that each of the files
// is of the same size.
func setProbabilitiesEqualSize(fsDescriptor *FileSystemNodeDescriptor, fsParentDescriptor *FileSystemNodeDescriptor) float32 {
	if len(fsDescriptor.Children) > 0 {
		var childrenPercent float32 = 0
		for i := range fsDescriptor.Children {
			childrenPercent += setProbabilitiesEqualSize(&fsDescriptor.Children[i], fsParentDescriptor)
		}
		fsDescriptor.Percent = childrenPercent
		return childrenPercent
	} else {
		filePercent := float32(1 / float64(fsParentDescriptor.NumFiles) * 100)
		fsDescriptor.Percent = filePercent
		return filePercent
	}
}

// setProbabilities calculates percentage of possibility of being randomly chosen
// for each of the nodes of a FileSystemDescriptor graph taking into consideration
// the different file sizes. Returns the total percentage assigned to the children.
func setProbabilities(fsDescriptor *FileSystemNodeDescriptor, fsParentDescriptor *FileSystemNodeDescriptor) float32 {
	if len(fsDescriptor.Children) > 0 {
		var childrenPercent float32 = 0
		for i := range fsDescriptor.Children {
			childrenPercent += setProbabilities(&fsDescriptor.Children[i], fsParentDescriptor)
		}
		fsDescriptor.Percent = childrenPercent
		return childrenPercent
	} else {
		if fsDescriptor.Percent == 0 {
			distributableAmount, distributablePercent := findDistributableAmountAndProbability(fsDescriptor, fsParentDescriptor)
			filePercent := float32(float64(fsDescriptor.Table.NumberOfStrings) / float64(distributableAmount) * float64(distributablePercent))
			fsDescriptor.Percent = filePercent
			return filePercent
		} else {
			return fsDescriptor.Percent
		}
	}
}

func findDistributableAmountAndProbability(fsDescriptor *FileSystemNodeDescriptor, fsParentDescriptor *FileSystemNodeDescriptor) (uint64, float32) {
	current := fsDescriptor
	for {
		if current.Percent > 0 {
			return fsParentDescriptor.NumEntries - fsParentDescriptor.UndefinedNumEntries, current.Percent
		}

		current = current.Parent
		if current == nil ||
			current.Parent == nil { // We do not want to process the parent as it always has 100% and that is not used defined
			break
		}
	}

	// If we reach this point, means that we should use the undefined percentage
	return fsParentDescriptor.UndefinedNumEntries, fsParentDescriptor.UndefinedChildrenPercent
}

func calculateUndefinedProbability(rootFsDescriptor *FileSystemNodeDescriptor) {
	var undefinedPercent float32 = 100
	var undefinedEntries uint64 = 0

	for i := range rootFsDescriptor.Children {
		undefinedPercent -= rootFsDescriptor.Children[i].Percent
		if rootFsDescriptor.Children[i].Percent == 0 {
			undefinedEntries += rootFsDescriptor.Children[i].NumEntries
		}
	}

	rootFsDescriptor.UndefinedChildrenPercent = undefinedPercent
	rootFsDescriptor.UndefinedNumEntries = undefinedEntries
}
