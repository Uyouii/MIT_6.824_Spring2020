#coding: utf8
import json
from base64 import b64decode
from Crypto.Cipher import AES
from Crypto.Util.Padding import unpad
import hmac
import hashlib

def AesDecode():
    raw_iv = b64decode("XyvzF/qCpDGwW+uTLW/CGg==")
    raw_encrypt = b64decode("+cPt3TfR/k11nIxqoYoZn9YUMpwVpGykoqAANfTX3wX6cSKg2wskHNRBBw1zP4q8c9XLmcA5/s22zhVei1fu7ZID8iujScwjVC3oRXW/ICYf95Gmvuw40gdA5fCB08WKqIB/osIuHPU7iXpIUTpl9KoX68OUZF2g/a9tePxyx2XbT4eDpJbpaWhdbXqQGW4bBueFVk7OYLn7xv6oxbxzNw==")
    key = b64decode("4h+xLrWKiGRcsYi2h7vNrA==")

    cipher = AES.new(key, AES.MODE_CBC, raw_iv)
    pt = unpad(cipher.decrypt(raw_encrypt), AES.block_size)

    print(pt)

def VoipHmacSha256():
    value_list = ['test', 'group2', "wx20afc706a711eefc"]
    value_list.sort()
    msg = ''
    for value in value_list:
        msg = msg + value
    print (msg)
    key = "bxLdikRXVbTPdHSM05e5u5sUoXNKd8-41ZO3MhKoyN5OfkWITDGgnr2fwJ0m9E8NYzWKVZvdVtaUgWvsdshFKA"
    signature = hmac.new(bytes(key , 'utf-8'),  bytes(msg , 'utf-8'), digestmod = hashlib.sha256).hexdigest()
    print (signature)

if __name__ == "__main__":

    VoipHmacSha256()