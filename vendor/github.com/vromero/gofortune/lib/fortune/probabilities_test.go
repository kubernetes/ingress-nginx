package fortune

import (
	"strconv"
	"testing"
)

func TestCalculateUndefinedPercents(t *testing.T) {

	secondLevelFile1 := FileSystemNodeDescriptor{
		NumEntries: 20,
	}

	secondLevelFile2 := FileSystemNodeDescriptor{
		NumEntries: 20,
	}

	firstLevelDir := FileSystemNodeDescriptor{
		Children:   []FileSystemNodeDescriptor{secondLevelFile1, secondLevelFile2},
		NumEntries: 40,
	}

	firstLevelFile := FileSystemNodeDescriptor{
		Percent:    33,
		NumEntries: 100,
	}

	root := FileSystemNodeDescriptor{
		Children: []FileSystemNodeDescriptor{firstLevelDir, firstLevelFile},
	}

	calculateUndefinedProbability(&root)
	if root.UndefinedChildrenPercent != 67 {
		t.Error("UndefinedChildrenPercent expected 67 got " + strconv.Itoa(int(root.UndefinedChildrenPercent)))
	}

	if root.UndefinedNumEntries != 40 {
		t.Error("UndefinedNumEntries expected 40 got " + strconv.Itoa(int(root.UndefinedNumEntries)))
	}
}
