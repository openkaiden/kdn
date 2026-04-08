Output demos/05_config_projects.gif

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

Type `mv ~/.kdn/config/projects.json ~/.kdn/config/projects.json.orig && echo -n '{"":{"environment":[{"name":"GH_TOKEN","secret":"my-github-token"}],"mounts":[{"host":"$HOME/.gitconfig","target":"$HOME/.gitconfig"}]}}' | json_pp > ~/.kdn/config/projects.json && clear`
Enter
Sleep 100ms

Show

Type "cat ~/.kdn/config/projects.json"
Enter
Sleep 5000ms

Type "echo my-super-github-secret | podman secret create my-github-token -"
Enter
Wait+Screen@5000ms /[0-9]/
Sleep 1000ms

Type "kdn init"
Enter
Wait+Screen@60000ms /Container image built/
Sleep 1000ms

Type "kdn terminal demo bash"
Enter
Sleep 2000ms

Type "echo $GH_TOKEN"
Enter 
Wait+Screen@1000ms /my-super-github-secret/
Sleep 2000ms

Type "mount | grep .gitconfig"
Enter
Wait+Screen@1000ms /gitconfig/
Sleep 5000ms

Hide

Type "exit"
Enter
Sleep 1000ms

Type "podman secret rm my-github-token"
Enter 
Sleep 100ms

Type "mv ~/.kdn/config/projects.json.orig ~/.kdn/config/projects.json"
Enter
Sleep 100ms

Type "kdn remove --force demo"
Enter
Sleep 25000ms
