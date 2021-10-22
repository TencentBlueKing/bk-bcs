# -*- coding: utf-8 -*-
"""
Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community
Edition) available.
Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://opensource.org/licenses/MIT

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.

python3版本
"""

from base64 import b64decode, b64encode

import requests
from Crypto.Cipher import DES
from Crypto.Hash import SHA
from cryptography.hazmat.backends import default_backend
from cryptography.hazmat.primitives import serialization


class BKDHException(Exception):
    pass


def _get_des_key(shared_key):
    key_bytes = []
    for b in shared_key[:8]:
        b = ord(b) if isinstance(b, str) else b
        _b = (b & 0xFE) | (
            (
                (
                    (b >> 1)
                    ^ (b >> 2)  # noqa
                    ^ (b >> 3)  # noqa
                    ^ (b >> 4)  # noqa
                    ^ (b >> 5)  # noqa
                    ^ (b >> 6)  # noqa
                    ^ (b >> 7)  # noqa
                )
                ^ 0x01
            )
            & 0x01
        )
        key_b = chr(_b).encode('latin1', errors='replace')
        key_bytes.append(key_b)
    return b''.join(key_bytes)


def _unpad(padded_data, block_size):
    pdata_len = len(padded_data)
    if pdata_len % block_size:
        raise ValueError("input data is not padded.")

    padding_len = padded_data[-1]
    return padded_data[:-padding_len]


def load_public_key(der_public_key):
    return serialization.load_der_public_key(b64decode(der_public_key), default_backend())


def dump_public_key(public_key):
    return b64encode(
        public_key.public_bytes(serialization.Encoding.DER, serialization.PublicFormat.SubjectPublicKeyInfo)
    )


def get_dh_parameters():
    # DHParameters 的数据存储在一个叫 dh_cdata 的指针中，很难通过 Python 修改
    # 或组装，所以这里直接通过一个合法的PublicKey读取其DHParameters得来
    public_key = load_public_key(
        "MEcwLQYJKoZIhvcNAQMBMCACExZWAhV0cUBBckkh" "WWg0c0IIBYcCBRI0VniQAgIAgAMWAAITDr8ZFFIy" "VMOTN83kE2zINJo/HQ=="
    )
    return public_key.parameters()


def decrypt(ciphertext, private_key, public_kb):
    shared_key = private_key.exchange(public_kb)
    key = _get_des_key(shared_key)
    des = DES.new(key, mode=DES.MODE_ECB)
    padded_data = des.decrypt(ciphertext)
    return _unpad(padded_data, DES.block_size)


def get_keypair():
    dh_parameters = get_dh_parameters()
    private_key = dh_parameters.generate_private_key()
    return private_key, private_key.public_key()


def request_api(url, public_ka, data_key, sha_key):
    r = requests.get(url, params={"publicKey": dump_public_key(public_ka)})
    try:
        data = r.json()
    except Exception:
        raise BKDHException("invalid response, json expected.")
    else:
        if 'data' in data:
            data = data['data']

    try:
        r.raise_for_status()
    except Exception:
        raise BKDHException("request failed: %r" % data.get('message'))
    return (load_public_key(data['publicKey']), b64decode(data[data_key]), data[sha_key])


def shortcuts(url, data_key, sha_key):
    private_key, public_ka = get_keypair()

    for i in range(3):
        public_kb, ciphertext, checksum = request_api(url, public_ka, data_key, sha_key)
        value = decrypt(ciphertext, private_key, public_kb)

        # validate sha1 checksum
        sha1 = SHA.new()
        sha1.update(value)
        if sha1.hexdigest() == checksum:
            return value.decode("utf-8")
    else:
        raise BKDHException("decrypt response failed.")
