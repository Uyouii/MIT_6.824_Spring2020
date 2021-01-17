#coding: utf8
import json
from base64 import b64decode
from Crypto.Cipher import AES
from Crypto.Util.Padding import unpad
import hmac
import hashlib

def AesDecode():
    raw_iv = b64decode("rI7IxmApxjLIZBnZRlVo0w==")
    raw_encrypt = b64decode("64NNvId913SBItQGZHLOJ3xvSIkmQvQ3z4aL7aFDroErdW4Og4Xmmqfcyh75vpVv")
    key = b64decode("IU4nr8OLkTo0n+btL0+oWQ==")

    cipher = AES.new(key, AES.MODE_CBC, raw_iv)
    pt = unpad(cipher.decrypt(raw_encrypt), AES.block_size)

    print(pt)

def HmacSha256_2(key_value_list, sig_key):
    value_list = [key + "=" + str(value) for key, value in key_value_list.items()]
    value_list.sort()
    msg = '&'.join(value_list)
    print (msg)
    signature = hmac.new(bytes(sig_key , 'utf-8'),  bytes(msg , 'utf-8'), digestmod = hashlib.sha256).hexdigest()
    print (signature)

def HmacSha256(value_list, key):
    value_list.sort()
    msg = ''.join(value_list)
    print (msg)
    signature = hmac.new(bytes(key , 'utf-8'),  bytes(msg , 'utf-8'), digestmod = hashlib.sha256).hexdigest()
    print (signature)

def Sha1(value, key):
    sha1 = hashlib.sha1()
    sha1.update((value + key).encode('utf-8'))
    print (sha1.hexdigest())

if __name__ == "__main__":
    Sha1("{\"nickName\":\"黄土地\",\"gender\":0,\"language\":\"zh_CN\",\"city\":\"\",\"province\":\"\",\"country\":\"\",\"avatarUrl\":\"https://thirdwx.qlogo.cn/mmopen/vi_32/DibyUCjaIR0E4JUViaWeqq2Hibs0ES9GPibFnLl2Sic4oPAo68OJMdtNkvF799HB96zmpHzcLBkZ8qNkJPdEPZsoCdA/132\"}", "0AajSXR9rjTkMv8pMLgdIA==")
    # value = ''
    # key = "wIeGXE93uSQjbpvdOSNf2g=="
    # signature = hmac.new(bytes(key , 'utf-8'),  bytes(value , 'utf-8'), digestmod = hashlib.sha256).hexdigest()
    # print (signature)
    # jump_wxa_key_value_list = {
    #     'appid': 'wx827225356b689e24',
    #     'timestamp' : '111',
    #     'nonce_str' : 'test',
    #     'client_ip' : '14.17.22.34',
    #     'wxa_appid' : 'wx20afc706a711eefc',
    #     'path' : '/',
    # }
    # share_wxa_key_value_list = {
    #     'appid': 'wx827225356b689e24',
    #     'timestamp' : '111',
    #     'nonce_str' : 'test',
    #     'client_ip' : '14.17.22.34',
    #     'title': "",
    #     'description':"",
    #     'thumb_img_url':'',
    #     'web_page_url':'',
    #     'wxa_appid' : 'wx20afc706a711eefc',
    #     'path' : '/',
    #     'hd_image_url':'',
    #     'with_share_ticket': 0,
    #     'wxa_type':0,

    # }
    # key = "Fi7Z6JLPS9CG87P63nXXdQ=="
    # msg = '{"kv_list":[{"key":"1","value":0}]}'
    # signature = hmac.new(bytes(key , 'utf-8'),  bytes(msg , 'utf-8'), digestmod = hashlib.sha256).hexdigest()
    # print ('key:', key)
    # print ('data:', msg)
    # print ('sig:', signature)