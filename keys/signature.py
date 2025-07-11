import base64
import requests
import time
import dotenv
import os
from cryptography.hazmat.primitives.serialization import load_pem_private_key
from cryptography.hazmat.primitives.asymmetric import ed25519
from typing import cast

dotenv.load_dotenv()

# Set up authentication
API_KEY=os.getenv('BINANCE_SPOT_TESTNET_API_KEY')
PRIVATE_KEY_PATH=os.getenv('ED25519_PRIVKEY_FILENAME')

if PRIVATE_KEY_PATH is None or PRIVATE_KEY_PATH == '':
    raise ValueError('PRIVATE_KEY_PATH is not set')

# Load the private key.
# In this example the key is expected to be stored without encryption,
# but we recommend using a strong password for improved security.
with open(PRIVATE_KEY_PATH, 'rb') as f:
    private_key = cast(ed25519.Ed25519PrivateKey, load_pem_private_key(data=f.read(), password=None))

response = requests.get('https://testnet.binance.vision/api/v3/ticker/price', params={'symbol': 'BTCUSDT'})
market_price = float(response.json()['price'])
print(f"Current market price: {market_price} USDT")

# Set up the request parameters
params = {
    'symbol':       'BTCUSDT',
    'side':         'SELL',
    'type':         'LIMIT',
    'timeInForce':  'GTC',
    'quantity':     '0.0001',
    'price':        '111351.12',
}

# Timestamp the request
timestamp = int(time.time() * 1000) # UNIX timestamp in milliseconds
params['timestamp'] = str(timestamp)

# Sign the request
payload = '&'.join([f'{param}={value}' for param, value in params.items()])

# Sign with ed25519 private key
signature = base64.b64encode(private_key.sign(payload.encode('ASCII'))).decode('ascii')
params['signature'] = signature

# Send the request
headers = {
    'X-MBX-APIKEY': API_KEY,
}
response = requests.post(
    'https://testnet.binance.vision/api/v3/order',
    headers=headers,
    data=params,
)
print(response.json())