package main

import "github.com/sirupsen/logrus"

func main() {
	arr := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	logrus.Info(arr)

	logrus.Info(arr[1:5])

	// copy rest of the array
	copy(arr[2:], arr[2+1:]) // dst: 3, 4..., start: 4, 5...
	//arr[len(arr)-1] = nil
	arr = arr[:len(arr)-1]

	logrus.Info(arr)

	// replace with last element
	arr[3] = arr[len(arr)-1]
	//arr[len(arr)-1] = nil
	arr = arr[:len(arr)-1]

	logrus.Info(arr)
}
