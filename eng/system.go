package eng

type System interface {
	Add() *Object
	Get(id EntityId) (ptr *Object, index int)
	Remove(index int)
	Reset()

	Update(dt float64)
	Draw(alpha float64)
}
