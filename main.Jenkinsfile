pipeline {
    agent { label 'agent' }
    environment {
        REGISTRY  = "192.168.56.10:5000"
        SSH_USER  = "vagrant"
        WORKER_IP = "192.168.56.22"
    }
    stages {        
        stage('Prepare Environment') {
            steps {
                script {
                    echo 'Pulling... ' + env.GIT_BRANCH
                    def BRANCH_NAME = env.GIT_BRANCH
                    env.SAFE_BRANCH = BRANCH_NAME.replaceAll('/', '-')                    
                    def branchKey = BRANCH_NAME.tokenize('/')[1]

                    // Define worker map
                    def workerMap = [
                        "prod1"     : "192.168.56.21",                        
                    ]
                    
                    if (BRANCH_NAME == "origin/main") {
                        echo "Pipeline running for Main branch."
                        env.WORKER_IP = workerMap['main']
                        env.WORKER_NAME = "main"
                    }
                    
                    echo "Branch ${BRANCH_NAME}, using safe tag ${env.SAFE_BRANCH}, deploying to worker ${env.WORKER_NAME} (${env.WORKER_IP})"
                }
            }
        }

        stage('Echo on Agent') {
            steps {
                sh 'echo "Running on agent: $(hostname)"'
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
                    input message: "Deploy to production?", ok: "Approve Deployment"
                }
            }
        }

        stage('Deploy on Worker') {
            when {
                expression { return currentBuild.result == null || currentBuild.result == 'SUCCESS' }
            }
            steps {
                script {
                    echo "Deploying to worker: ${WORKER_IP}"
                }
                sh """
                sshpass -p "vagrant" ssh -o StrictHostKeyChecking=no ${SSH_USER}@${WORKER_IP} \\
                'if [ \$(docker ps -q) ]; then docker stop \$(docker ps -a -q); fi && docker run -d --network host --volume appData:/etc/todos --pull=always ${REGISTRY}/${SAFE_BRANCH}:${IMAGE_VERSION}'
                """
            }
        }
    }
}
