# Copyright 2016 The Prometheus Authors
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

include Makefile.common

DOCKER_IMAGE_NAME ?= graphite-exporter


# FIXME(matthiasr): this should be part of the test suite, but it does not
# finish at least on TravisCI.
end-to-end-test: build
	@echo ">> running end-to-end test"
	@bash end-to-end-test.sh
