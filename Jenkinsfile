pipeline {
    agent { label 'agent' }
    environment {
        REGISTRY  = "192.168.56.10:5000"
        SSH_USER  = "vagrant"
        // BRANCH_NAME = "feature/jenkins" // Temporarily set this manually
        // SAFE_BRANCH = "feature-jenkins"
    }
    stages {        
        stage('Prepare Environment') {
            steps {
                script {
                    if (env.BRANCH_NAME == 'origin/main') {
                        echo "Skipping build for main branch"
                        currentBuild.result = 'ABORTED'
                        error("Main branch is managed separately. Stopping the pipeline.")
                    }
                    echo 'Pulling... ' + env.GIT_BRANCH
                    // Replace "/" with "-" so the tag is Docker-friendly.
                    def BRANCH_NAME = env.GIT_BRANCH
                    env.SAFE_BRANCH = BRANCH_NAME.replaceAll('/', '-')                    
                    // Extract a key from the branch name.
                    // For example, "feature/jenkins" gives branchKey = "feature"
                    def branchKey = BRANCH_NAME.tokenize('/')[1]
                    // Map branch keys to worker IPs.
                    echo "branch name ${BRANCH_NAME}"
                    echo "branch key ${branchKey}"
                    def workerMap = [
                        "dev"     : "192.168.56.23",
                        "feature" : "192.168.56.24"
                    ]
                    
                    if (workerMap.containsKey(branchKey)) {
                        env.WORKER_IP   = workerMap[branchKey]
                        env.WORKER_NAME = branchKey 
                    } else {
                        echo "No worker configured for branch ${BRANCH_NAME} using feature as default"
                        env.WORKER_IP   = workerMap['feature']
                        env.WORKER_NAME = "feature" 
                    }
                    
                    echo "For branch ${BRANCH_NAME}, using safe tag ${env.SAFE_BRANCH} and deploying to worker ${env.WORKER_NAME} (${env.WORKER_IP})"
                }
            }
        }
        
        // Confirm that the job is running on the agent.
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
        
        stage('Deploy on Worker') {
            when {
                expression { return currentBuild.result == null || currentBuild.result == 'SUCCESS' }
            }
            steps {
                sh """
                sshpass -p "vagrant" ssh -o StrictHostKeyChecking=no ${SSH_USER}@${WORKER_IP} \\
                'if [ \$(docker ps -q) ]; then docker stop \$(docker ps -a -q); fi && docker run -d --network host --volume appData:/etc/todos --pull=always ${REGISTRY}/${SAFE_BRANCH}:${IMAGE_VERSION}'
                """
            }
        }
    }
}
