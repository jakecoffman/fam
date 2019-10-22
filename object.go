package fam

var objectId int

func GetObjectId() int {
	objectId++
	return objectId
}
