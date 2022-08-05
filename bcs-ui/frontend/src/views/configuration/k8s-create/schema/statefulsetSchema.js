/*
* Tencent is pleased to support the open source community by making
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
*
* Copyright (C) 2021 THL A29 Limited, a Tencent company.  All rights reserved.
*
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) is licensed under the MIT License.
*
* License for 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition):
*
* ---------------------------------------------------
* Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
* documentation files (the "Software"), to deal in the Software without restriction, including without limitation
* the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and
* to permit persons to whom the Software is furnished to do so, subject to the following conditions:
*
* The above copyright notice and this permission notice shall be included in all copies or substantial portions of
* the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO
* THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF
* CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
* IN THE SOFTWARE.
*/
const schema = {
  type: 'object',
  properties: {
    metadata: {
      type: 'object',
      properties: {
        name: {
          type: 'string',
        },
      },
    },
    restartPolicy: {
      type: 'object',
      properties: {
        policy: { type: 'string', enum: ['Never', 'Always', 'OnFailure'] },
        interval: { type: 'number', minimum: 0 },
        backoff: { type: 'number', minimum: 0 },
        maxtimes: { type: 'number', minimum: 0 },
      },
    },
    killPolicy: {
      type: 'object',
      properties: {
        gracePeriod: { type: 'number', minimum: 0 },
      },
    },
    constraint: {
      type: 'object',
      properties: {
        intersectionItem: {
          type: 'array',
          items: {
            type: 'object',
            properties: {
              unionData: {
                type: 'array',
                items: {
                  type: 'object',
                  properties: {
                    name: { type: 'string', enum: ['hostname', 'InnerIp'] },
                    operate: {
                      type: 'string',
                      enum: ['UNIQUE', 'MAXPER', 'CLUSTER', 'GROUPBY', 'LIKE', 'UNLIKE'],
                    },
                    type: { type: 'number', minimum: 3, maximum: 4 },
                  },
                },
              },
            },
          },
        },
      },
    },
    spec: {
      type: 'object',
      properties: {
        instance: { type: 'number', minimum: 1 },
        template: {
          type: 'object',
          properties: {
            spec: {
              type: 'object',
              properties: {
                networkMode: {
                  type: 'string',
                  enum: ['CUSTOM', 'BRIDGE', 'HOST', 'USER', 'NONE'],
                },
                networkType: { type: 'string', enum: ['cni', 'cnm'] },
                containers: {
                  type: 'array',
                  items: {
                    type: 'object',
                    properties: {
                      type: { type: 'string', enum: ['MESOS'] },
                      image: { type: 'string' },
                      imagePullPolicy: { type: 'string', enum: ['Always', 'IfNotPresent'] },
                      privileged: { type: 'boolean' },
                      resources: {
                        type: 'object',
                        properties: {
                          limits: {
                            type: 'object',
                            properties: {
                              cpu: [
                                { type: 'number', minimum: 0 },
                                { type: 'string' },
                              ],
                              memory: [
                                { type: 'number', minimum: 0 },
                                { type: 'string' },
                              ],
                            },
                          },
                        },
                      },
                      volumes: {
                        type: 'array',
                        items: {
                          type: 'object',
                          properties: {
                            name: { type: 'string' },
                            volume: {
                              type: 'object',
                              properties: {
                                hostPath: {
                                  type: 'string',
                                },
                                mountPath: {
                                  type: 'string',
                                },
                              },
                            },
                          },
                        },
                      },
                      healthChecks: {
                        type: 'array',
                        items: {
                          type: 'object',
                          properties: {
                            type: {
                              type: 'string',
                              enum: [
                                '',
                                'HTTP',
                                'TCP',
                                'COMMAND',
                                'REMOTE_HTTP',
                                'REMOTE_TCP',
                              ],
                            },
                            delaySeconds: { type: 'number', minimum: 0 },
                            intervalSeconds: { type: 'number', minimum: 0 },
                            timeoutSeconds: { type: 'number', minimum: 0 },
                            consecutiveFailures: { type: 'number', minimum: 0 },
                            gracePeriodSeconds: { type: 'number', minimum: 0 },
                            command: {
                              type: 'object',
                              properties: {
                                value: { type: 'string' },
                              },

                            },
                            tcp: {
                              type: 'object',
                              properties: {
                                port: {
                                  oneOf: [
                                    { type: 'number' },
                                    { type: 'string' },
                                  ],
                                },
                                portName: {
                                  oneOf: [
                                    {
                                      type: 'string',
                                    },
                                  ],
                                },
                              },
                            },
                            http: {
                              type: 'object',
                              properties: {
                                port: {
                                  oneOf: [
                                    { type: 'number' },
                                    { type: 'string' },
                                  ],
                                },
                                portName: {
                                  oneOf: [
                                    {
                                      type: 'string',
                                    },
                                  ],
                                },
                                scheme: { type: 'string' },
                                path: { type: 'string' },
                              },
                            },
                          },
                        },
                      },
                      ports: {
                        type: 'array',
                        items: {
                          type: 'object',
                          properties: {
                            protocol: {
                              type: 'string',
                              enum: ['HTTP', 'TCP', 'UDP', ''],
                            },
                            name: {
                              oneOf: [
                                {
                                  type: 'string',
                                },
                              ],
                            },
                            hostPort: {
                              oneOf: [
                                {
                                  type: 'string',
                                },
                                {
                                  type: 'number',
                                  minimum: 31000,
                                  maximum: 32000,
                                },
                                {
                                  type: 'number',
                                  minimum: 0,
                                  maximum: 0,
                                },
                              ],
                            },
                            containerPort: {
                              oneOf: [
                                {
                                  type: 'string',
                                },
                                {
                                  type: 'number',
                                  minimum: 1,
                                  maximum: 65535,
                                },
                              ],
                            },
                          },
                        },
                      },
                      command: { type: 'string' },
                      args: { type: 'array' },
                      env: {
                        type: 'array',
                        items: {
                          type: 'object',
                          properties: {
                            name: {
                              type: 'string',
                            },
                            value: {
                              type: 'string',
                            },
                          },
                        },
                      },
                      parameters: {
                        type: 'array',
                        items: {
                          type: 'object',
                          properties: {
                            key: { type: 'string', minLength: 1 },
                            value: { type: 'string', minLength: 1 },
                          },
                        },
                      },
                    },
                  },
                },
              },
            },
          },
        },
      },
    },
  },
};

export {
  schema,
};
