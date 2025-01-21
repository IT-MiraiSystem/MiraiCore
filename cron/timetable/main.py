import os
import json
import requests
from datetime import datetime
AppDir = os.path.dirname(os.path.abspath(__file__))
with open(os.path.join(AppDir + 'config/config.json')) as f:
    config = json.load(f)
with open(os.path.join(AppDir+"config/FirebaseConfig.json"))as f:
    firebaseConfig = json.load(f)

if __name__ == "__main__":
    BaseURL=f"http://api{config["location"]}/"
    