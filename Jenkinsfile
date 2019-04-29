pipeline{

 agent any

   stages{
       stage("Testing Stage"){
           steps{
               sh "cd /home/shubham/go/src/jenkins/workspace/Jenkins-Pipeline"
               sh "go test"

                }
             }
       stage("Building Application"){
           steps{
               sh "/home/shubham/go/src/jenkins/workspace/Jenkins-Pipeline"
               sh "go build"
              }

           }
         }
       }
