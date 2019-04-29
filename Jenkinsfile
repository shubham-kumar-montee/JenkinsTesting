pipeline{

 agent any

    tools {
        go 'Go 1.10.3'
    }

   stages{
       stage("Testing Stage"){
           steps{
               sh "go test"

                }
             }
       stage("Building Application"){
           steps{
               sh "go build"
              }

           }
         }
       }
