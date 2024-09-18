package utils

import "github.com/google/uuid"


func HeightToIndex(height uint32) uint32 {
	return height
}

func RandID() string {
	return uuid.New().String()
}