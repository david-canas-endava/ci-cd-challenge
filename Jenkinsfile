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
                    // Replace "/" with "-" so the tag is Docker-friendly.
                    BRANCH_NAME = env.BRANCH_NAME
                    env.SAFE_BRANCH = BRANCH_NAME.replaceAll('/', '-')                    
                    // Extract a key from the branch name.
                    // For example, "feature/jenkins" gives branchKey = "feature"
                    def branchKey = BRANCH_NAME.tokenize('/')[0]
                    // Map branch keys to worker IPs.
                    echo "branch name ${BRANCH_NAME}"
                    echo "branch key ${branchKey}"
                    def workerMap = [
                        "prod1"   : "192.168.56.21",
                        "prod2"   : "192.168.56.22",
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
        
        // Check out the code.
        // In a multibranch pipeline, Jenkins checks out the branch automatically.
        stage('Checkout Code') {
            steps {
                checkout scm
            }
        }
        
        // Build the Docker image, tag it, and push it.
        stage('Build and Push Docker Image') {
            steps {
                // Ensure the target directory exists (if you intend to use /app)
                sh 'mkdir -p /app'
                
                // Use the /app directory. If your checkout already happens in /app, adjust accordingly.
                dir('/home/jenkins_agent/workspace/githubPipeline/app') {
                    // Build an image tagged as "app"
                    sh 'docker build -t app .'
                    
                    // Tag the built image with the registry and the branch-derived tag.
                    sh "docker tag app ${REGISTRY}/${SAFE_BRANCH}"
                    
                    // Push the image to your local registry.
                    sh "docker push ${REGISTRY}/${SAFE_BRANCH}"
                }
            }
        }
        
        // SSH into the appropriate worker machine and run the container.
        stage('Deploy on Worker') {
            steps {
                // The following command will SSH into the worker machine and run the container.
                // The container will be named "app_${WORKER_NAME}" and will run the image we just pushed.
                // Note: Make sure the agent’s SSH keys and authorization are correctly set up.
                sh """
                sshpass -p "vagrant" ssh -o StrictHostKeyChecking=no ${SSH_USER}@${WORKER_IP} \\
                'docker run -d --network host --volume appData:/etc/todos ${REGISTRY}/${SAFE_BRANCH}'
                """
            }
        }
    }
}
