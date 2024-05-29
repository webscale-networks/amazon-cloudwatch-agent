/**
 * Jenkinsfile existing in a project is seen by jenkins and used to build a
 * project. For documentation, see:
 * https://jenkins.io/doc/book/pipeline/jenkinsfile/. For variables that can
 * be used, see
 * https://jenkins.webscalenetworks.com/job/product/job/control/pipeline-syntax/globals.
 */

properties([
  // only keep 180 days of builds to prevent out-of-memory when accessing build
  buildDiscarder(
    logRotator(
      artifactDaysToKeepStr: '',
      artifactNumToKeepStr: '',
      daysToKeepStr: '180',
      numToKeepStr: '',
    ),
  ),
])

/*
 * Compute an environment from a set of properties in a map. This is
 * a separate function because it has to be annotated as non-cps in
 * order to be able to use the standard groovy collect method. Not
 * sure why.
 */
@com.cloudbees.groovy.cps.NonCPS
def environment(props) {
  props.collect{ k, v -> k + '=' + v }
}


def isMerge = false
def commitEmail
def commitSubject
def success = false
def interrupted = false

try {
  node {
    currentBuild.displayName = 'branch ' + env.BRANCH_NAME + ' ' + currentBuild.displayName
    /* Checks out the main source tree */
    stage('scm') {
      def scmVars = checkout scm
      isMerge = sh(returnStdout: true, script: "git log -n 1 --format=%P").trim().split().size() > 1
      commitEmail = sh(returnStdout: true, script: "git log -n 1 --pretty=format:%ae").trim()
      commitSubject = sh(returnStdout: true, script: "git log -n 1 --pretty=format:%s").trim()
    }

    stage('build') {
      sh('make release')
    }

    stage('publish') {
      //def version = "${sh('git describe --tags --abbrev=0')}-${env.BRANCH_NAME}"
      //def svn = "file:///repo/amazon-cloudwatch-agent/${version}"
      //sh("mkdir ${version}; cp build/bin/linux_amd64/amazon-cloudwatch-agent ${version}/")
      //sh("svn import --no-ignore --quiet -m 'Update to version ${version}' ${version} ${svn}")
    }
  }
}
catch (InterruptedException e) {
  interrupted = true
  throw e
}
finally {
  if (!interrupted) {
    def state
    def color
    def icon
    if (success) {
      state = 'success'
      color = '#53B636'
      icon = ':white_check_mark:'
    } else {
      state = 'failed'
      color = '#E45F2F'
      icon = ':crash:'
    }
    if (state != 'success' || !isMerge) {
      def message = "${icon} <${env.BUILD_URL}|control ${env.BUILD_DISPLAY_NAME}> ${state}"
      if (commitEmail) {
        message += "\n${commitEmail} ${commitSubject}"
      }
      slackSend(
        channel: '#engineering-botlover',
        color: color,
        message: message,
      )
    }
  }
}
