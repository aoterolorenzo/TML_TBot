package controllers

import (
	"fmt"
	"testing"
)

func TestTMLLineUpController_CompareLineUps(t1 *testing.T) {
	want := "----- Friday 21 July 2023 ----\n\nCage Stage\n- Patrick Mason se mueve de las 16:00 - 18:00 a las 26:00\n\nTerra Solis Stage\n- Eliminado Mosoo (13:30 - 15:00)\n\nAtmosphere Stage\n- Añadido Adam Beyer a las (23:00 - 01:00)"
	c := TMLLineUpController{}

	initialLineUp, err := c.Retrieve()
	if err != nil {
		t1.Errorf(err.Error())
	}
	updatedLineUp, err := c.Retrieve()
	if err != nil {
		t1.Errorf(err.Error())
	}

	delete(initialLineUp["Friday 21 July 2023"]["Atmosphere"], "Adam Beyer")
	initialLineUp["Friday 21 July 2023"]["Cage"]["Patrick Mason"] = "26:00"

	delete(updatedLineUp["Friday 21 July 2023"]["Terra Solis"], "Mosoo")

	got, err := c.CompareLineUps(initialLineUp, updatedLineUp)
	if err != nil {
		t1.Errorf(err.Error())
	}
	fmt.Println()
	fmt.Println(got)
	fmt.Println()

	if got != want {
		return
	}
}
