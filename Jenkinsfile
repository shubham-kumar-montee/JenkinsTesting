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
     post {

 always {
            step([$class: 'Publisher', reportFilenamePattern: 'testng.xml'])
        }
       success{
             publishHTML([allowMissing: false, alwaysLinkToLastBuild: false, keepAll: false, reportDir: '/home/shubham/GIT-WORKSPACE/JenkinsTesting/HTML Report', reportFiles: 'index.html', reportName: 'HTML Report', reportTitles: 'REPORT HAI BHAI'])

             emailext body: 'all the stages has passed', subject: 'testing', to: 'kantusjee123123@gmail.com'
                  }
            }
   
       }
