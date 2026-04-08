Output demos/04_config_local.gif

Set Shell "bash"
Set FontSize 20
Set Width 1600
Set Height 600
Set Padding 16
Set Framerate 15
Set TypingSpeed 30ms

Hide

Type "export KDN_DEFAULT_RUNTIME=podman"
Enter
Sleep 100ms

Type "export KDN_DEFAULT_AGENT=claude"
Enter
Sleep 100ms

Type "export KDN_INIT_AUTO_START=1"
Enter
Sleep 100ms


Type `mkdir .kaiden && echo -n '{"environment":[{"name":"KEY1","value":"val1"}],"mounts":[{"host":"$HOME/.gitconfig","target":"$HOME/.gitconfig"}]}' | json_pp > .kaiden/workspace.json && clear`
Enter
Sleep 100ms

Show

Type "cat .kaiden/workspace.json"
Enter
Sleep 5000ms

Type "kdn init"
Enter
Wait+Screen@25000ms /Container started/
Sleep 1000ms

Type "kdn terminal demo bash"
Enter
Sleep 1000ms

Type "echo $KEY1"
Enter 
Wait+Screen@1000ms /val1/
Sleep 3000ms

Type "mount | grep .gitconfig"
Enter
Sleep 5000ms

Hide

Type "exit"
Enter
Sleep 1000ms

Type "rm .kaiden/workspace.json && rmdir .kaiden"
Enter
Sleep 100ms

Type "kdn remove --force demo"
Enter
Sleep 25000ms
