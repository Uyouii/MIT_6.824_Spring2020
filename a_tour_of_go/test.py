#coding: utf8
import json
from base64 import b64decode
from Crypto.Cipher import AES
from Crypto.Util.Padding import unpad


raw_iv = b64decode("LHq6A65ml2gyllAMi+F/8g==")
raw_encrypt = b64decode("eRKUAT9EiMHz8xyhwB044q48ArRt9rgVUEyb+9mdDRgxPgrQbupBti9KOwFvyWedEM+bQ0iYPMkj6prrHNeuM7O1l2KrMqLXDJgRZfC6kk5FhOBtRp7DmZvkMQVZyq5+U9nFAi8fgsG5up4+xdYbB2i4FSD6+MXgsVSvwJDJ75wE2en5diFi2VYS9s/btlrQYuAnAQ/eTUJ+EkMMpUsuLXSyMMrNxWOJ541W0txOK31EMJTdLPZmnOSef/H9k9/cVsR/XVJb5b7xoN7+8niWFVEVJPSW10M9TC4SByn7+7tkWvdbjNvnA9PvnZhlcGkOFforCUtino4gLBZ1+Q5D0AWlHRcRRVGjZfHGAFd+7IOSqdbKZayilHSWEn5Iso9DBCuc4g5yrjvicE0UN6K2eKvOJoE9rALXSt5c1SZSL4rJ5ikNNFztz9ROhrPqx82GuBI9juV8FIVfEnL9SlB/OBIvv8oS1wt2G5dKov6bfJuuA2xUgseryMgD/ziMTynV")
key = b64decode("2IHa2SzkVd/TkNEO0dbAwg==")

cipher = AES.new(key, AES.MODE_CBC, raw_iv)
pt = unpad(cipher.decrypt(raw_encrypt), AES.block_size)

print(pt)
