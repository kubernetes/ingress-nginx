@Library('libpipelines@master') _


hose {
    EMAIL = 'eos@stratio.com'
    BUILDTOOL_IMAGE = 'stratio/ingress-nginx-builder:0.2.0'
    BUILDTOOL = 'make'
    DEVTIMEOUT = 30
    DEPLOYONPRS = true
    ANCHORE_TEST = true
    VERSIONING_TYPE = 'stratioVersion-3-3'
    UPSTREAM_VERSION = '1.2.0'

    DEV = { config ->
        doPackage(config)
        doDocker(conf: config, dockerfile: "rootfs/Dockerfile.stratio")
        doHelmChart(conf: config, helmTarget: "chart")
    }
}
