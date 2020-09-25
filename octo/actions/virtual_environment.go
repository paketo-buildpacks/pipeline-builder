/*
 * Copyright 2018-2020 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package actions

type VirtualEnvironment string

const (
	WindowsLatest VirtualEnvironment = "windows-latest"
	Windows2019   VirtualEnvironment = "windows-2019"
	Ubuntu2004    VirtualEnvironment = "ubuntu-20.04"
	UbuntuLatest  VirtualEnvironment = "ubuntu-latest"
	Ubuntu1804    VirtualEnvironment = "ubuntu-18.04"
	Ubuntu1604    VirtualEnvironment = "ubuntu-16.04"
	MacOSLatest   VirtualEnvironment = "macos-latest"
	MacOS1015     VirtualEnvironment = "macos-10.15"
	SelfHosted    VirtualEnvironment = "self-hosted"
)
