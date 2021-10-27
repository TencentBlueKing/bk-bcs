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
Package scheduler provides scheduler main logic implements.

Transactions
There are followed transactions to do applications or taskgroups operation:
LAUCH: create and launch an application from version definition
DELETE: delete application
UPDATE: update application
SCALE: scale up or scale down application's instances
RESCHEDULE: reschedule taskgroup when it is fail or required by API

Service
When applications are running, sometimes they are binded to some services, and need to export to services,
Service Manager is implemented to do application bind and export, it watches followed events:
Taskgroup Add
Taskgroup Delete
Taskgroup Update
Service Add
Service Update
Service delete

Status Report
When tasks run on slave machine, the status will reported by mesos slave, the report message is processed by function StatusReport

Health Check Report
If a running taskgroup is configured to do health check, the health-check result will reported by healthy module, the messeages are processed by HealthyReport

Deployment related functions
The deployments' rollingupdate is implemented by using application transactions, refer to function DeploymentCheck

DataChecker
DataChecker is responsable for dirty or error data in ZK
refer to DataCheckManage

Message
Message is used to send message to executor, just as localfile, signal ...
*/
package scheduler
