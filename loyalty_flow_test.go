package main

import (
	"fmt"
	"testing"
	
)

func  TestCheck(t *testing.T) {
	fmt.Println("Entering TestInit")
	name:=maincheck()
        if(name!="shubham"){
     t.Errorf("returned value not the same")
}
	
	
}


