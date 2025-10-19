pipeline {
  agent any

  environment {
    // Homebrew path for Apple Silicon; adjust if needed
    PATH = "/opt/homebrew/bin:${env.PATH}"
    GOMODCACHE = "${env.WORKSPACE}/.gomodcache"
  }

  options {
    timestamps()
  }

  stages {
    stage('Checkout') {
      steps {
        checkout scm
      }
    }

    stage('Setup') {
      steps {
        sh 'which go || true'
        sh 'go version || true'
        sh 'go env'
        sh 'go mod download'
      }
    }

    stage('Test') {
      steps {
        sh 'go test ./... -v -race -coverprofile=coverage.out'
      }
    }
  }

  post {
    always {
      archiveArtifacts artifacts: 'coverage.out', fingerprint: true
    }
    success {
      echo 'Tests passed!'
    }
    failure {
      echo 'Tests failed — check the console output.'
    }
  }
}
