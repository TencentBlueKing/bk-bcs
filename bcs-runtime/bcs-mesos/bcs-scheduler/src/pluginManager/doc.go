/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

/*
Package pluginManager provides dymamic getting mesos slave attributes implements.

Scheduler support for the use of plugins to get mesos slave attributes.
It is mainly applicable to the acquisition of dynamic attributes, example for container ip resources,
net flow.

The types of plugin are mainly including dynamic, executable.
User can implement specific plugin based on their own scenarios.

	//mesos slave attribute plugin's names
	pluginsNames := []string{"ip-resources","net-flow"}

	pluginer,err := NewPluginManager(pluginsNames)
	if err != nil {
		//...
		return err
	}

	//get mesos slave's dynamic attributes
	attrs,err := pluginer.GetHostAttributes(para)
	if err != nil {
		//...
		return err
	}

	//set dynamic attributes to mesos slave
	//...
*/
package pluginManager
