pipeline{

 agent any

    tools {
        go 'Go 1.10.3'
    }

   stages{
       stage("Testing Stage"){
           steps{
               sh "go test"
  publishHTML([allowMissing: false, alwaysLinkToLastBuild: false, keepAll: false, reportDir: '/home/shubham/GIT-WORKSPACE/JenkinsTesting', reportFiles: 'report.html', reportName: 'HTML Report', reportTitles: 'Report'])

                }
             }
       stage("Building Application"){
           steps{
               sh "go build"
  publishHTML([allowMissing: false, alwaysLinkToLastBuild: false, keepAll: false, reportDir: '/home/shubham/GIT-WORKSPACE/JenkinsTesting', reportFiles: 'report.html', reportName: 'HTML Report', reportTitles: 'Report'])
              }
           }
         }
     post {
       success{
                publishHTML([allowMissing: false, alwaysLinkToLastBuild: false, keepAll: false, reportDir: '/home/shubham/GIT-WORKSPACE/JenkinsTesting', reportFiles: 'report.html', reportName: 'HTML Report', reportTitles: 'Report'])
                emailext body: 'all the stages has passed', subject: 'testing', to: 'kantusjee123123@gmail.com'
             }


          }
       }
