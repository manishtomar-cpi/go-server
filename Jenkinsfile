pipeline {
  agent any

  environment {
    // Make Homebrew Go visible to Jenkins on macOS
    PATH = "/opt/homebrew/bin:${env.PATH}"
    GOMODCACHE = "${env.WORKSPACE}/.gomodcache"
  }

  options { timestamps() }

  stages {
    stage('Checkout') {
      steps { checkout scm }
    }

    stage('Setup') {
      steps {
        sh 'go version'
        sh 'go mod download'
      }
    }

    stage('Unit tests') {
      steps {
        sh 'go test ./... -v -race -coverprofile=coverage.out'
      }
      post {
        always {
          archiveArtifacts artifacts: 'coverage.out', fingerprint: true
          sh 'go tool cover -html=coverage.out -o coverage.html || true'
          archiveArtifacts artifacts: 'coverage.html', fingerprint: true
        }
      }
    }
  }

  post {
    success { echo 'Tests passed' }
    failure { echo ' Tests failed — see Console Output' }
  }
}
