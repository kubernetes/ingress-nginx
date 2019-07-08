#!/usr/bin/env python

# Copyright 2018 The Kubernetes Authors.
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

"""
Convert the output of `nginx-ingress-controller --help` to
a Markdown table.
"""

import re
import sys

assert sys.version_info[0] == 3, 'This script requires Python 3'

data = sys.stdin.read()
data = data.replace('\t', ' ' * 8)  # Expand tabs
data = data.replace('\n' + (' ' * 8 * 2), ' ')  # Unwrap lines

print('''
| Argument | Description |
|----------|-------------|
'''.rstrip())

for arg_m in re.finditer('^\s+(-.+?)\s{2,}(.+)$', data, flags=re.MULTILINE):
	arg, description = arg_m.groups()
	print('| `{arg}` | {description} |'.format(
		arg=arg.replace(', ', '`, `'),
		description=description,
	))
