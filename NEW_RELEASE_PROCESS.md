# Semi-Automated Release Process 

1. Update TAG
2. Cloud Build 
3. k8s.io PR 
4. git pull origin main 
5. git checkout -b $RELEASE_VERSION 
6. mage release:newrelease $RELEASE_VERSION 
7. Wait for PR