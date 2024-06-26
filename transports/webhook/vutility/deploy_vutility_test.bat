@ECHO OFF

:choice
set /P c=deploy to TEST[Y/N]?
if /I "%c%" EQU "Y" goto :deploy
if /I "%c%" EQU "N" goto :exit
goto :choice

:deploy

echo "Deploying"

call go mod vendor
echo "Mod Vendor"
call gcloud config set project *****-test
call gcloud run deploy *your-cloud-run* --source ./  --region "*your-preferred-region*"

:exit
echo "exiting"
