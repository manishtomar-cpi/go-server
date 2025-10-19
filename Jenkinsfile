pipeline {
  agent any

  tools {
    go 'Go1.22'   // must match the name of your Go tool configured in Jenkins
  }

  options {
    timestamps()
    ansiColor('xterm')
  }

  environment {
    GOMODCACHE = "${env.WORKSPACE}/.gomodcache"
  }

  stages {
    stage('Checkout') {
      steps { checkout scm }
    }
    stage('Setup') {
      steps {
        sh 'go version'
        sh 'go env'
        sh 'go mod download'
      }
    }
    stage('Test') {
      steps {
        sh 'go test ./... -v -race -coverprofile=coverage.out'
      }
      post {
        always {
          archiveArtifacts artifacts: 'coverage.out', fingerprint: true
        }
      }
    }
  }

  post {
    success { echo 'Tests passed!' }
    failure { echo 'Tests failed — check console log.' }
  }
}
