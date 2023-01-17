@Library('libpipelines@master') _


hose {
    EMAIL = 'eos@stratio.com'
    BUILDTOOL_IMAGE = 'golang:1.19'
    BUILDTOOL = 'make'
    DEVTIMEOUT = 30
    VERSIONING_TYPE = 'stratioVersion-3-3'
    UPSTREAM_VERSION = '1.2.1'
    ANCHORE_TEST = true
    DEPLOYONPRS = true

    DEV = { config ->
        doPackage(config)
        doDocker(conf: config, dockerfile: "rootfs/Dockerfile.stratio")
        doHelmChart(conf: config, helmTarget: "chart")
    }
}
