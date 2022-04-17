# SpotifyBackupper
Automatically saves a snapshot of your spotify playlists as backup

# Usage
 1. Download the the latest release and unzip the folder.
 
 2. After download you need to login to your spotify account and setup the config file.
   To do so, simply run the get_token.exe file in the get_token folder.
   It will give you a url where you have to login to your spotify account.

     After you've done so, you will have to edit the newly created config.json file in the Release folder.
     Add your client ID and client Secret, which you need from the spotify web api, to the config.json file.
     Finally enter the "backuppath" in the json file. It should point to a .json file where the backup will be safed.
 
 3. After configuring is done, you can let the SpotifyBackupper.exe run on device startup or invoke it
 manually from time to time to backup your spotify playlists.
