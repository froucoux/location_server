package quadtree

import (
	"testing"
	"rand"
	"time"
	"strconv"
	"container/vector"
)

const treeMaxSize = 10000

type dim struct {
	width, height float64
}

const dups = 10

// Slice of random dimensions for creating quad trees
var dims = []dim{
	dim{10, 10},
	dim{1, 2},
	dim{100, 300},
	dim{20.4, 35.6},
	dim{1e10, 500.000001},
}

type point struct {
	x, y float64
}

var testRand = rand.New(rand.NewSource(time.Nanoseconds()))

// Tests whether a newly built quadtree has the correct dimensions
func TestEmpty(t *testing.T) {
	for _, r := range dims {
		testEmpty(r.width, r.height, t)
	}
}

func testEmpty(width, height float64, t *testing.T) {
	var tree = NewArrayTree(0, width, 0, height, treeMaxSize)
	view := tree.View()
	switch false {
	case view.lx == 0:
		t.Errorf("View has %f for lx, expecting 0", view.lx)
	case view.rx == width:
		t.Error("View has %f for rx, expecting %f", view.rx, width)
	case view.ty == 0:
		t.Error("View has %f for ty, expecting 0", view.ty)
	case view.by == height:
		t.Error("View has %f for by, expecting 1", view.by, height)
	}
}

// Test that we can insert a single element into the tree and then retrieve it
func TestOneElement(t *testing.T) {
	for _, r := range dims {
		tree := NewArrayTree(0, r.width, 0, r.height, treeMaxSize)
		testOneElement(tree, t)
	}
}

func testOneElement(tree T, t *testing.T) {
	x, y := randomPosition(tree.View())
	tree.Insert(x, y, "test")
	fun, results := SimpleSurvey()
	tree.Survey([]*View{tree.View()}, fun)
	if results.Len() != 1 || "test" != results.At(0) {
		t.Errorf("Failed to find required element at (%f,%f), in tree \n%v", x, y, tree)
	}
}

// Test that if we add 5 elements into a single quadrant of a fresh tree
// We can successfully retrieve those elements. This test is tied to
// the implementation detail that a quadrant with 5 elements will 
// over-load a single leaf and must rearrange itself to fit the 5th 
// element in.
func TestFullLeaf(t *testing.T) {
	for _, r := range dims {
		w := r.width
		h := r.height
		v := OrigView(w, h)
		v1, v2, v3, v4 := v.quarters()
		testFullLeaf(NewArrayTree(0, w, 0, h, treeMaxSize), v1, "v1", t)
		testFullLeaf(NewArrayTree(0, w, 0, h, treeMaxSize), v2, "v2", t)
		testFullLeaf(NewArrayTree(0, w, 0, h, treeMaxSize), v3, "v3", t)
		testFullLeaf(NewArrayTree(0, w, 0, h, treeMaxSize), v4, "v4", t)
	}
}

func testFullLeaf(tree T, v *View, msg string, t *testing.T) {
	for i := 0; i < 5; i++ {
		x, y := randomPosition(v)
		name := "test" + strconv.Itoa(i)
		tree.Insert(x, y, name)
	}
	vt := tree.View()
	fun, results := SimpleSurvey()
	tree.Survey([]*View{vt}, fun)
	if results.Len() != 5 {
		t.Error(msg, "Inserted 5 elements into a fresh quadtree and retrieved only ", results.Len())
	}
}

// Tests that we can add a large number of random elements to a tree
// and create random views for collecting from the populated tree.
func TestScatter(t *testing.T) {
	for _, r := range dims {
		testScatter(NewArrayTree(0, r.width, 0, r.height, treeMaxSize), t)
	}
}

func testScatter(tree T, t *testing.T) {
	ps := fillView(tree.View(), 2)
	for i, p := range ps {
		tree.Insert(p.x, p.y, "test"+strconv.Itoa(i))
	}
	for i := 0; i < 1; i++ {
		sv := subView(tree.View())
		var count int
		for _, v := range ps {
			if sv.contains(v.x, v.y) {
				count++
			}
		}
		fun, results := SimpleSurvey()
		tree.Survey([]*View{sv}, fun)
		if count != results.Len() {
			t.Errorf("Failed to retrieve %d elements in scatter test, found %d instead", count, results.Len())
		}
	}
}

// Tests that we can add multiple elements to the same location
// and still retrieve all elements, including duplicates, using 
// randomly generated views.
func testScatterDup(tree T, t *testing.T) {
	ps := fillView(tree.View(), 1000)
	for _, p := range ps {
		for i := 0; i < dups; i++ {
			tree.Insert(p.x, p.y, "test_"+strconv.Itoa(i))
		}
	}
	for i := 0; i < 1000; i++ {
		sv := subView(tree.View())
		var count int
		for _, v := range ps {
			if sv.contains(v.x, v.y) {
				count++
			}
		}
		fun, results := SimpleSurvey()
		tree.Survey([]*View{sv}, fun)
		if count*dups != results.Len() {
			t.Errorf("Failed to retrieve %d elements in duplicate scatter test, found only %d", count*3, results.Len())
		}
	}
}

// Tests that when we
// 1: Add a single element to an empty tree
// 2: Remove that element from the tree
// We get
// 1: The single element is the only element in the deleted list
// 2: The tree no longer contains any elements
func TestSimpleAddDelete(t *testing.T) {
	for _, d := range dims {
		testAddDelete(d, t)
		testAddDeleteDup(d, t)
		testAddDeleteMulti(d, t)
	}
}

// Tests a very limited deletion scenario. Here we will insert every element in 'insert' into the tree at a
// single random point. Then we will delete every element in delete from the tree. 
// If exact == true then the view used to delete covers eactly the insertion point. Otherwise, it covers the
// entire tree.
// We assert that every element of delete has been deleted from the tree (testDelete)
// We assert that every element in insert but not in delete is still in the tree (testSurvey)
// errPrfx is used to distinguish the error messages from different tests using this method.
func testDeleteSimple(tree T, insert, delete []interface{}, exact bool, errPrfx string, t *testing.T) {
	x, y := randomPosition(tree.View())
	for _, e := range insert {
		tree.Insert(x, y, e)
	}
	expCol := new(vector.Vector)
OUTER_LOOP:
	for _, i := range insert {
		for _, d := range delete {
			if i == d {
				continue OUTER_LOOP
			}
		}
		expCol.Push(i)
	}
	expDel := new(vector.Vector)
	for _, d := range delete {
		expDel.Push(d)
	}
	pred, deleted := makeDelClosure(delete)
	delView := tree.View()
	if exact {
		delView = NewViewP(x, x, y, y)
	}
	testDelete(tree, delView, pred, deleted, expDel, t, errPrfx)
	fun, collected := SimpleSurvey()
	testSurvey(tree, tree.View(), fun, collected, expCol, t, errPrfx)
}

// Add element delete everything from the tree.
func testAddDelete(d dim, t *testing.T) {
	elem := "test"
	testDeleteSimple(NewArrayTree(0, d.width, 0, d.height, treeMaxSize), []interface{}{elem}, []interface{}{elem}, false, "Simple Global Delete", t)
	testDeleteSimple(NewArrayTree(0, d.width, 0, d.height, treeMaxSize), []interface{}{elem}, []interface{}{elem}, true, "Simple Exact Delete", t)
}

// Add two elements, delete one element from entire tree
func testAddDeleteDup(d dim, t *testing.T) {
	elem := "test"
	elemII := "testII"
	testDeleteSimple(NewArrayTree(0, d.width, 0, d.height, treeMaxSize), []interface{}{elem, elemII}, []interface{}{elem}, false, "Simple Gobal Delete Take One Of Two", t)
	testDeleteSimple(NewArrayTree(0, d.width, 0, d.height, treeMaxSize), []interface{}{elem, elemII}, []interface{}{elem}, false, "Simple Exact Delete Take One Of Two", t)
}

// Add two elements, delete both from entire tree
func testAddDeleteMulti(d dim, t *testing.T) {
	elem := "test"
	elemII := "testII"
	//testDeleteSimple(NewArrayTree(0, d.width, 0, d.height, treeMaxSize), []interface{}{elem, elemII}, []interface{}{elem, elemII}, false, "Simple Global Delete Take Two Of Two", t)
	testDeleteSimple(NewArrayTree(0, d.width, 0, d.height, treeMaxSize), []interface{}{elem, elemII}, []interface{}{elem, elemII}, true, "Simple Exact Delete Take Two Of Two", t)
}

func TestScatterDelete(t *testing.T) {
	for _, r := range dims {
		for i := 0; i < 10; i++ {
			testScatterDelete(NewArrayTree(0, r.width, 0, r.height, treeMaxSize), t)
			testScatterDeleteMulti(NewArrayTree(0, r.width, 0, r.height, treeMaxSize), t)
		}
	}
}

func testScatterDelete(tree T, t *testing.T) {
	name := "test"
	pointNum := 1000
	ps := fillView(tree.View(), pointNum)
	for i, p := range ps {
		tree.Insert(p.x, p.y, name+strconv.Itoa(i))
	}
	delView := subView(tree.View())
	expDel := new(vector.Vector)
	expCol := new(vector.Vector)
	for i, p := range ps {
		if delView.contains(p.x, p.y) {
			expDel.Push(name + strconv.Itoa(i))
		} else {
			expCol.Push(name + strconv.Itoa(i))
		}
	}
	pred, deleted := CollectingDelete()
	testDelete(tree, delView, pred, deleted, expDel, t, "Scatter Insert and Delete Under Area")
	fun, collected := SimpleSurvey()
	testSurvey(tree, tree.View(), fun, collected, expCol, t, "Scatter Insert and Delete Under Area")
}

func testScatterDeleteMulti(tree T, t *testing.T) {
	name := "test"
	pointNum := 1000
	points := fillView(tree.View(), pointNum)
	for i, p := range points {
		for d := 0; d < dups; d++ {
			tree.Insert(p.x, p.y, name+strconv.Itoa(i)+"_"+strconv.Itoa(d))
		}
	}
	delView := subView(tree.View())
	expDel := new(vector.Vector)
	expCol := new(vector.Vector)
	for i, p := range points {
		if delView.contains(p.x, p.y) {
			for d := 0; d < dups; d++ {
				expDel.Push(name + strconv.Itoa(i) + "_" + strconv.Itoa(d))
			}
		} else {
			for d := 0; d < dups; d++ {
				expCol.Push(name + strconv.Itoa(i) + "_" + strconv.Itoa(d))
			}
		}
	}
	pred, deleted := CollectingDelete()
	testDelete(tree, delView, pred, deleted, expDel, t, "Scatter Insert and Delete Under Area With Three Elements Per Location")
	fun, results := SimpleSurvey()
	testSurvey(tree, tree.View(), fun, results, expCol, t, "Scatter Insert and Delete Under Area With Three Elements Per Location")
}

func testDelete(tree T, view *View, pred func(x, y float64, e interface{}) bool, deleted, expDel *vector.Vector, t *testing.T, errPfx string) {
	tree.Delete(view, pred)
	if deleted.Len() != expDel.Len() {
		t.Errorf("%s: Expecting %v deleted element(s), found %v", errPfx, expDel.Len(), deleted.Len())
	}
OUTER_LOOP:
	for i := 0; i < expDel.Len(); i++ {
		for j := 0; j < deleted.Len(); j++ {
			expVal := expDel.At(i)
			delVal := deleted.At(j)
			if expVal == delVal {
				continue OUTER_LOOP
			}
		}
		t.Errorf("%s: Expecting to find %v in deleted vector, was not found", errPfx, expDel.At(i))
	}
}

func testSurvey(tree T, view *View, fun func(x, y float64, e interface{}), collected, expCol *vector.Vector, t *testing.T, errPfx string) {
	tree.Survey([]*View{view}, fun)
	if collected.Len() != expCol.Len() {
		t.Errorf("%s: Expecting %v collected element(s), found %v", errPfx, expCol.Len(), collected.Len())
	}
	/* This code checks that every expected element is present
		   In practice this is too slow - disabled
	OUTER_LOOP:
		for i := 0; i < expCol.Len(); i++ {
			expVal := expCol.At(i)
			for j := 0; j < collected.Len(); j++ {
				colVal := collected.At(j)
				if expVal == colVal {
					continue OUTER_LOOP
				}
			}
			t.Errorf("%s: Expecting to find %v in collected vector, was not found", errPfx, expCol.At(i))
		}
	*/
}

// Creates a closure which deletes all elements which are present in elem
// Returns the closure plus a vector.Vector into which deleted elements are accumulated
func makeDelClosure(elems []interface{}) (pred func(x, y float64, e interface{}) bool, deleted *vector.Vector) {
	deleted = new(vector.Vector)
	pred = func(x, y float64, e interface{}) bool {
		for i := range elems {
			if e == elems[i] {
				deleted.Push(e)
				return true
			}
		}
		return false
	}
	return
}

func randomPosition(v *View) (x, y float64) {
	x = testRand.Float64()*(v.rx-v.lx) + v.lx
	y = testRand.Float64()*(v.by-v.ty) + v.ty
	return
}

func fillView(v *View, c int) []point {
	ps := make([]point, c)
	for i := 0; i < c; i++ {
		x, y := randomPosition(v)
		ps[i] = point{x: x, y: y}
	}
	return ps
}

func subView(v *View) *View {
	lx := testRand.Float64()*(v.rx-v.lx) + v.lx
	rx := testRand.Float64()*(v.rx-lx) + lx
	ty := testRand.Float64()*(v.by-v.ty) + v.ty
	by := testRand.Float64()*(v.by-ty) + ty
	return NewViewP(lx, rx, ty, by)
}
