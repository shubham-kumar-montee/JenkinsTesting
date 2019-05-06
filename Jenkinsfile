pipeline{

 agent any

    tools {
        go 'Go 1.10.3'
    }

   stages{
       stage("Testing Stage"){
           steps{
               sh "go get -u github.com/jstemmer/go-junit-report"
               sh "go test -v 2>&1 | go-junit-report > report.xml"
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
            step([$class: 'Publisher', reportFilenamePattern: 'testing.xml'])
        }
       success{
             publishHTML([allowMissing: false, alwaysLinkToLastBuild: false, keepAll: false, reportDir: '/home/shubham/GIT-WORKSPACE/JenkinsTesting/HTML Report', reportFiles: 'index.html', reportName: 'HTML Report', reportTitles: 'REPORT HAI BHAI'])

             emailext body: 'all the stages has passed', subject: 'testing', to: 'kantusjee123123@gmail.com'
                  }
            }
   
       }
