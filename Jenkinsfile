node {
    stage  ('Load shared libraries') {
        def version='v4.5'
        fileLoader.withGit('https://git.aurora.skead.no/scm/ao/aurora-pipeline-scripts.git', version) {
            maven = fileLoader.load('maven/maven')
            git = fileLoader.load('git/git')
            go = fileLoader.load('go/go')
        }
    }

  stage('Checkout') {
    checkout scm
  }

  stage('Test and coverage'){
    go.buildGoWithJenkinsSh()
  }

  stage('Deploy to Nexus'){
    def isMaster = env.BRANCH_NAME == 'master'
    
    def REPO_ID = isMaster ? 'releases' : 'snapshots'
    def REPO_URL = 'https://aurora/nexus/content/repositories/' + REPO_ID
    
    def version = git.getTagFromCommit()

    if (isMaster){
      if (!git.tagExists("v${version}")) {
        error "Commit is not tagged. Aborting build."
      }
    }

    def deployOpts = '-Durl=' + REPO_URL + 
        ' -DrepositoryId=' + REPO_ID + 
        ' -DgroupId=ske.aurora.openshift -DartifactId=radish -Dversion=' + version + 
        ' -Dpackaging=tar.gz -DgeneratePom=true -Dfile=bin/amd64/radish.tar.gz'

    maven.setMavenVersion('Maven 3')
    maven.run('deploy:deploy-file', deployOpts)

  }

}
