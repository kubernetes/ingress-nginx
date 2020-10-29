#!/usr/bin/env python3

# Copyright 2020 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import sys

from ruamel.yaml import YAML

yaml=YAML()
yaml.indent(mapping=2, sequence=4, offset=2)

for manifest in yaml.load_all(sys.stdin.read()):
    if manifest:
        # helm template does not have support for namespace declaration
        if ('metadata' in manifest and 'namespace' not in manifest['metadata']
                and manifest['kind'] != 'Namespace' and manifest['kind'] != 'ClusterRole'
                and manifest['kind'] != 'ClusterRoleBinding' and manifest['kind'] != 'ValidatingWebhookConfiguration'):
            manifest['metadata']['namespace'] = sys.argv[1]

        # respect existing replicas definition
        if 'spec' in manifest and 'replicas' in manifest['spec']:
            del manifest['spec']['replicas']

        print('---')
        yaml.dump(manifest, sys.stdout)
