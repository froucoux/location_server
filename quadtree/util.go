package quadtree

import (
	"container/vector"
)

//
//	A Simple quadtree collector which will push every element into col
//
func SimpleSurvey() (fun func(x, y float64, e interface{}), col *vector.Vector) {
	col = new(vector.Vector)
	fun = func(x, y float64, e interface{}) {
		col.Push(e)
	}
	return
}

//
//	A Simple quadtree delete function which indicates that every element given to it should be deleted
//
func SimpleDelete() (pred func(x, y float64, e interface{}) bool) {
	pred = func(x, y float64, e interface{}) bool {
		return true
	}
	return
}

//
//	A quadtree delete function which indicates that every element given to it should be deleted.
//	Additionally each element deleted will be pushed into col
//
func CollectingDelete() (pred func(x, y float64, e interface{}) bool, col *vector.Vector) {
	col = new(vector.Vector)
	pred = func(x, y float64, e interface{}) bool {
		col.Push(e)
		return true
	}
	return
}

// 
//	Determines if a point lies inside at least one of a slice of *View
//
func contains(vs []*View, x, y float64) bool {
	for _, v := range vs {
		if v.contains(x, y) {
			return true
		}
	}
	return false
}

//
//	Determines if a view overlaps at least one of a slice of *View
//
func overlaps(vs []*View, oV *View) bool {
	for _, v := range vs {
		if oV.overlaps(v) {
			return true
		}
	}
	return false
}

func max(f1, f2 float64) float64 {
	if f1 > f2 {
		return f1
	}
	return f2
}

func min(f1, f2 float64) float64 {
	if f1 < f2 {
		return f1
	}
	return f2
}
