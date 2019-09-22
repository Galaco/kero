package bsp

import "github.com/galaco/bsp"

func LoadBSP(filename string) (*bsp.Bsp, error) {
	return bsp.ReadFromFile(filename)
}
