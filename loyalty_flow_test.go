package main

import (
	"fmt"
	"testing"
	
)

func checkMaincheck(t *testing.T) {
	fmt.Println("Entering TestInit")
	name:=maincheck()
        if(name!="shubham"){
     t.Errorf("Sum was incorrect, got: %d, want: %d.", total, 10)
}
	
	
}


