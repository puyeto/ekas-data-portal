pipeline {
    environment {
      DOCKER = credentials('docker_hub')
    }
    agent any
        stages {
            stage('Build') {
                parallel {
                    stage('Express Image') {
                        steps {
                            sh 'docker build -f Dockerfile \
                            -t omollo/ekas-ntsa-data-prod:latest .'
                        }
                    }                    
                }
                post {
                    failure {
                        echo 'This build has failed. See logs for details.'
                    }
                }
            }
            stage('Test') {
                steps {
                    echo 'This is the Testing Stage'
                }
            }
            stage('DEPLOY') {
                when {
                    branch 'master'  //only run these steps on the master branch
                }
                steps {
                    // sh 'docker swarm leave -f'
                    sh 'docker run -d -p 8082:8082 --rm --name ekas-data-portal ekas-ntsa-data-prod'
                    // sh 'docker swarm init --advertise-addr 159.89.134.228'
                    // sh 'docker stack deploy -c docker-compose.yml ekas-ntsa-data-prod'
                }

            }

            // stage('REPORTS') {
            //     steps {
            //         junit 'reports.xml'
            //         archiveArtifacts(artifacts: 'reports.xml', allowEmptyArchive: true)
            //         // archiveArtifacts(artifacts: 'ekas-ntsa-data-prod-golden.tar.gz', allowEmptyArchive: true)
            //     }
            // }

            stage('CLEAN-UP') {
                steps {
                    // sh 'docker stop ekas-ntsa-data-dev'
                    sh 'docker system prune -f'
                    deleteDir()
                }
            }
        }
    }