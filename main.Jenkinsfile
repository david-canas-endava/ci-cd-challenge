pipeline {
    agent { label 'agent' }
    environment {
        REGISTRY  = "192.168.56.10:5000"
        SSH_USER  = "vagrant"
        WORKER_IP = "192.168.56.21"
        SECOND_WORKER_IP = "192.168.56.22"
    }
    stages {        
        stage('Prepare Environment') {
            steps {
                script {
                    echo 'Pulling... ' + env.GIT_BRANCH
                    def BRANCH_NAME = env.GIT_BRANCH
                    env.SAFE_BRANCH = BRANCH_NAME.replaceAll('/', '-')                    
                    def branchKey = BRANCH_NAME.tokenize('/')[1]

                    def workerMap = [
                        "prod1"     : "192.168.56.21",
                        "prod2"     : "192.168.56.22",
                    ]
                    
                    echo "Branch ${BRANCH_NAME}, using safe tag ${env.SAFE_BRANCH}, deploying to worker ${env.WORKER_NAME} (${env.WORKER_IP})"

                    echo "WORKER_IP inside script: ${env.WORKER_IP}"
                    echo "SECOND_WORKER_IP inside script: ${env.SECOND_WORKER_IP}"
                }
            }
        }

        stage('Echo on Agent') {
            steps {
                sh 'echo "Running on agent: $(hostname)"'
                echo "WORKER_IP outside script: ${env.WORKER_IP}"
                echo "SECOND_WORKER_IP outside script: ${env.SECOND_WORKER_IP}"
            }
        }

        stage('Checkout Code') {
            steps {
                checkout scm
                script {
                    env.IMAGE_VERSION = sh(
                        script: '[ -s version ] && cat version || echo "latest"',
                        returnStdout: true
                    ).trim()
                }
            }
        }

        stage('Build Docker Image') {
            steps {
                sh 'mkdir -p /home/jenkins_agent/workspace/githubPipeline/app'
                dir('/home/jenkins_agent/workspace/githubPipeline/app') {
                    sh 'docker build -t app .'
                    sh "docker tag app ${REGISTRY}/${SAFE_BRANCH}:${IMAGE_VERSION}"
                    sh "if [ \$(docker ps -q) ]; then docker stop \$(docker ps -a -q); fi"
                    sh "docker run -d --network host --volume /etc/todos:/etc/todos ${REGISTRY}/${SAFE_BRANCH}:${IMAGE_VERSION}"
                }
            }
        }

        stage('Run Tests') {
            steps {
                dir('/home/jenkins_agent/workspace/githubPipeline/test') {
                    sh 'go mod download'
                    sh 'go test -v ./main_test.go'
                }
            }
        }

        stage('Push Docker Image') {
            when {
                expression { return currentBuild.result == null || currentBuild.result == 'SUCCESS' }
            }
            steps {
                sh "docker push ${REGISTRY}/${SAFE_BRANCH}:${IMAGE_VERSION}"
            }
        }

        stage('Approval for Deploy (Main Branch)') {            
            steps {
                script {
                    def userChoice = input(
                        message: "Select Deployment Type",
                        parameters: [
                            choice(name: 'DEPLOYMENT_TYPE', choices: ['Simple Deploy', 'No Deploy', 'Canary Deploy'], description: 'Choose deployment strategy')
                        ]
                    )

                    env.DEPLOYMENT_TYPE = userChoice
                }
            }
        }

        stage('Full Deployment') {
            when {
                expression { env.DEPLOYMENT_TYPE == 'Simple Deploy' }
            }
            steps {
                script {
                    echo "Deploying to Worker 1: ${WORKER_1}"
                }
                sh """
                sshpass -p "vagrant" ssh -o StrictHostKeyChecking=no ${SSH_USER}@${WORKER_1} \\
                'if [ \$(docker ps -q) ]; then docker stop \$(docker ps -a -q); fi && docker run -d --network host --volume appData:/etc/todos --pull=always ${REGISTRY}/${SAFE_BRANCH}:${IMAGE_VERSION}'
                """

                script {
                    echo "Deploying to Worker 2: ${WORKER_2}"
                }
                sh """
                sshpass -p "vagrant" ssh -o StrictHostKeyChecking=no ${SSH_USER}@${WORKER_2} \\
                'if [ \$(docker ps -q) ]; then docker stop \$(docker ps -a -q); fi && docker run -d --network host --volume appData:/etc/todos --pull=always ${REGISTRY}/${SAFE_BRANCH}:${IMAGE_VERSION}'
                """
            }
        }

        stage('Canary Deployment') {
            when {
                expression { env.DEPLOYMENT_TYPE == 'Canary Deploy' }
            }
            steps {
                script {
                    echo "Deploying canary to Worker 1: ${WORKER_1}"
                }
                sh """
                sshpass -p "vagrant" ssh -o StrictHostKeyChecking=no ${SSH_USER}@${WORKER_1} \\
                'if [ \$(docker ps -q) ]; then docker stop \$(docker ps -a -q); fi && docker run -d --network host --volume appData:/etc/todos --pull=always ${REGISTRY}/${SAFE_BRANCH}:${IMAGE_VERSION}'
                """
                
                // Placeholder for HTTP signal (not implemented yet)
                // sh "curl -X POST http://monitoring-service/canary-deployed"
            }
        }

        stage('Approval to Continue Canary Deployment') {
            when {
                expression { env.DEPLOYMENT_TYPE == 'Canary Deploy' }
            }
            steps {
                script {
                    def proceed = input(
                        message: "Canary deployment completed. Continue full deployment?",
                        parameters: [
                            choice(name: 'CONTINUE_DEPLOYMENT', choices: ['Yes', 'No'], description: 'Proceed with full deployment?')
                        ]
                    )

                    env.CONTINUE_DEPLOYMENT = proceed
                }
            }
        }

        stage('Canary Deployment part 2') {
            when {
                expression { env.DEPLOYMENT_TYPE == 'Canary Deploy' && env.CONTINUE_DEPLOYMENT == 'Yes' }
            }
            steps {
                script {
                    echo "Deploying canary to Worker 2: ${WORKER_2}"
                }
                sh """
                sshpass -p "vagrant" ssh -o StrictHostKeyChecking=no ${SSH_USER}@${WORKER_2} \\
                'if [ \$(docker ps -q) ]; then docker stop \$(docker ps -a -q); fi && docker run -d --network host --volume appData:/etc/todos --pull=always ${REGISTRY}/${SAFE_BRANCH}:${IMAGE_VERSION}'
                """
                
                // Placeholder for HTTP signal (not implemented yet)
                // sh "curl -X POST http://monitoring-service/canary-deployed"
            }
        }

        // stage('Deploy on Worker 1') {
        //     when {
        //         expression { return currentBuild.result == null || currentBuild.result == 'SUCCESS' }
        //     }
        //     steps {
        //         script {
        //             echo "Deploying to worker: ${env.WORKER_IP}"
        //         }
        //         sh """
        //         sshpass -p "vagrant" ssh -o StrictHostKeyChecking=no ${SSH_USER}@${env.WORKER_IP} \\
        //         'if [ \$(docker ps -q) ]; then docker stop \$(docker ps -a -q); fi && docker run -d --network host --volume appData:/etc/todos --pull=always ${REGISTRY}/${SAFE_BRANCH}:${IMAGE_VERSION}'
        //         """
        //     }
        // }
        // stage('Approval for Deploy (Main Branch)') {            
        //     steps {
        //         script {
        //             input message: "Deploy to production?", ok: "Approve Deployment"
        //         }
        //     }
        // }
        
        // stage('Deploy on Worker 2') {
        //     when {
        //         expression { return currentBuild.result == null || currentBuild.result == 'SUCCESS' }
        //     }
        //     steps {
        //         script {
        //             echo "Deploying to worker: ${env.WORKER_IP}"
        //         }
        //         sh """
        //         sshpass -p "vagrant" ssh -o StrictHostKeyChecking=no ${SSH_USER}@${env.SECOND_WORKER_IP} \\
        //         'if [ \$(docker ps -q) ]; then docker stop \$(docker ps -a -q); fi && docker run -d --network host --volume appData:/etc/todos --pull=always ${REGISTRY}/${SAFE_BRANCH}:${IMAGE_VERSION}'
        //         """
        //     }
        // }
    }
}
