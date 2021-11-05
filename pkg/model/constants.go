package model

// Constants for a rocket chat installation
const (
	MongodbComponentName          = "mongodb"
	MongodbTargetPort             = "mongodb"
	MongodbDefaultVersion         = "4.4.10"
	MongodbScriptPath             = "/scripts/setup.sh"
	MongodbStatefulSetSuffix      = "-mongodb"
	MongodbServiceSuffix          = "-mongodb-service"
	MongodbHeadlessServiceSuffix  = "-mongodb-service-headless"
	MongodbScriptsConfigmapSuffix = "-mongodb-scripts"
	MongodbVolumeSuffix           = "-datadir"
	MongodbAuthSecretSuffix       = "-mongodb-auth"

	RocketAdminSecretSuffix         = "-admin"
	RocketWebserverComponentName    = "webserver"
	RocketWebserverDefaultVersion   = "3.18"
	RocketWebserverDeploymentSuffix = "-rocketchat"
	RocketWebserverServiceSuffix    = "-rocketchat-service"
)

var (
	RocketWebserverUser     = int64(999)
	RocketWebserverGroup    = int64(999)
	MongodbScriptMode       = int32(0755)
	MongodbUser             = int64(1001)
	MongodbReadinessCommand = `# Run the proper check depending on the version
[[ $(mongo --version | grep "MongoDB shell") =~ ([0-9]+\.[0-9]+\.[0-9]+) ]] && VERSION=${BASH_REMATCH[1]}
. /opt/bitnami/scripts/libversion.sh
VERSION_MAJOR="$(get_sematic_version "$VERSION" 1)"
VERSION_MINOR="$(get_sematic_version "$VERSION" 2)"
VERSION_PATCH="$(get_sematic_version "$VERSION" 3)"
if [[ "$VERSION_MAJOR" -ge 4 ]] && [[ "$VERSION_MINOR" -ge 4 ]] && [[ "$VERSION_PATCH" -ge 2 ]]; then
    mongo --disableImplicitSessions $TLS_OPTIONS --eval 'db.hello().isWritablePrimary || db.hello().secondary' | grep -q 'true'
else
    mongo --disableImplicitSessions $TLS_OPTIONS --eval 'db.isMaster().ismaster || db.isMaster().secondary' | grep -q 'true'
fi`
	boolTrue = true
)
