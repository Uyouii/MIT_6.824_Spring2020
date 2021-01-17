#coding: utf8
import json
from base64 import b64decode
import hmac
import hashlib

def HmacSha256(key_value_list, sig_key):
    value_list = [key + "=" + str(value) for key, value in key_value_list.items()]
    value_list.sort()
    msg = '&'.join(value_list)
    print (msg)
    signature = hmac.new(bytes(sig_key , 'utf-8'),  bytes(msg , 'utf-8'), digestmod = hashlib.sha256).hexdigest()
    print (signature)

if __name__ == "__main__":
    value_list = {
        'appid': 'wx827225356b689e24',
        'timestamp' : '111',
        'nonce_str' : 'test',
        'wx_token' : 'f7484632123f182dc812a0b8729b8526',
        'wxa_appid' : 'wx20afc706a711eefc',
        'path' : '/',
    }
    key = "HoagFKDcsGMVCIY2vOjf9tufiH_WappS1bmxYvsikm1rrBCU3ijBavITsYs9eB13GpDark-IJTAhhWY-cEa2vw"
    HmacSha256(value_list, key)
