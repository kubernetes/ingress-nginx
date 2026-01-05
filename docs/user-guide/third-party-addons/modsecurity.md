# ModSecurity Web Application Firewall

ModSecurity is an open source, cross platform web application firewall (WAF) engine for Apache, IIS and Nginx that is developed by OWASP. It has a robust event-based programming language which provides protection from a range of attacks against web applications and allows for HTTP traffic monitoring, logging and real-time analysis - [https://www.modsecurity.org](https://www.modsecurity.org)

The [ModSecurity-nginx](https://github.com/owasp-modsecurity/ModSecurity-nginx) connector is the connection point between NGINX and libmodsecurity (ModSecurity v3).

The default ModSecurity configuration file is located in `/etc/nginx/modsecurity/modsecurity.conf`. This is the only file located in this directory and contains the default recommended configuration. Using a volume we can replace this file with the desired configuration.
To enable the ModSecurity feature we need to specify `enable-modsecurity: "true"` in the configuration configmap.

>__Note:__ the default configuration use detection only, because that minimizes the chances of post-installation disruption.
Due to the value of the setting [SecAuditLogType=Concurrent](https://github.com/owasp-modsecurity/ModSecurity/wiki/Reference-Manual-(v2.x)#secauditlogtype) the ModSecurity log is stored in multiple files inside the directory `/var/log/audit`.
The default `Serial` value in SecAuditLogType can impact performance.

The OWASP ModSecurity Core Rule Set (CRS) is a set of generic attack detection rules for use with ModSecurity or compatible web application firewalls. The CRS aims to protect web applications from a wide range of attacks, including the OWASP Top Ten, with a minimum of false alerts.
The directory `/etc/nginx/owasp-modsecurity-crs` contains the [OWASP ModSecurity Core Rule Set repository](https://github.com/coreruleset/coreruleset).
Using `enable-owasp-modsecurity-crs: "true"` we enable the use of the rules.

## Supported annotations

For more info on supported annotations, please see [annotations/#modsecurity](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#modsecurity)

## Example of using ModSecurity with plugins via the helm chart

Suppose you have a ConfigMap that contains the contents of the [nextcloud-rule-exclusions plugin](https://github.com/coreruleset/nextcloud-rule-exclusions-plugin/blob/main/plugins/nextcloud-rule-exclusions-before.conf) like this:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: modsecurity-plugins
data:
  empty-after.conf: |
    # no data
  empty-before.conf: |
    # no data
  empty-config.conf: |
    # no data
  nextcloud-rule-exclusions-before.conf:
    # this is just a snippet
    # find the full file at https://github.com/coreruleset/nextcloud-rule-exclusions-plugin
    #
    # [ File Manager ]
    # The web interface uploads files, and interacts with the user.
    SecRule REQUEST_FILENAME "@contains /remote.php/webdav" \
        "id:9508102,\
        phase:1,\
        pass,\
        t:none,\
        nolog,\
        ver:'nextcloud-rule-exclusions-plugin/1.2.0',\
        ctl:ruleRemoveById=920420,\
        ctl:ruleRemoveById=920440,\
        ctl:ruleRemoveById=941000-942999,\
        ctl:ruleRemoveById=951000-951999,\
        ctl:ruleRemoveById=953100-953130,\
        ctl:ruleRemoveByTag=attack-injection-php"
```

If you're using the helm chart, you can pass in the following parameters in your `values.yaml`:

```yaml
controller:
  config:
    # Enables Modsecurity
    enable-modsecurity: "true"

    # Update ModSecurity config and rules
    modsecurity-snippet: |
      # this enables the mod security nextcloud plugin
      Include /etc/nginx/owasp-modsecurity-crs/plugins/nextcloud-rule-exclusions-before.conf

      # this enables the default OWASP Core Rule Set
      Include /etc/nginx/owasp-modsecurity-crs/nginx-modsecurity.conf

      # Enable prevention mode. Options: DetectionOnly,On,Off (default is DetectionOnly)
      SecRuleEngine On

      # Enable scanning of the request body
      SecRequestBodyAccess On

      # Enable XML and JSON parsing
      SecRule REQUEST_HEADERS:Content-Type "(?:text|application(?:/soap\+|/)|application/xml)/" \
        "id:200000,phase:1,t:none,t:lowercase,pass,nolog,ctl:requestBodyProcessor=XML"

      SecRule REQUEST_HEADERS:Content-Type "application/json" \
        "id:200001,phase:1,t:none,t:lowercase,pass,nolog,ctl:requestBodyProcessor=JSON"

      # Reject if larger (we could also let it pass with ProcessPartial)
      SecRequestBodyLimitAction Reject

      # Send ModSecurity audit logs to the stdout (only for rejected requests)
      SecAuditLog /dev/stdout

      # format the logs in JSON
      SecAuditLogFormat JSON

      # could be On/Off/RelevantOnly
      SecAuditEngine RelevantOnly

  # Add a volume for the plugins directory
  extraVolumes:
    - name: plugins
      configMap:
        name: modsecurity-plugins

  # override the /etc/nginx/enable-owasp-modsecurity-crs/plugins with your ConfigMap
  extraVolumeMounts:
    - name: plugins
      mountPath: /etc/nginx/owasp-modsecurity-crs/plugins
```
