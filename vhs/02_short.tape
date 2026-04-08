Output demos/02_short.gif

Set Shell "bash"
Set FontSize 20
Set Width 1600
Set Height 600
Set Padding 16
Set Framerate 15
Set TypingSpeed 30ms

Type "export KDN_DEFAULT_RUNTIME=podman"
Enter
Sleep 2000ms

Type "export KDN_DEFAULT_AGENT=claude"
Enter
Sleep 2000ms

Type "export KDN_INIT_AUTO_START=1"
Enter
Sleep 2000ms

Type "kdn init"
Enter
Wait+Screen@5000ms /Container started/
Sleep 5000ms

Type "kdn list"
Enter
Wait+Screen@5000ms /demo/
Sleep 5000ms

Type "clear"
Enter
Sleep 1500ms

Type "kdn terminal demo"
Enter
Sleep 5000ms

Type "/exit"
Enter
Sleep 2000ms

Type "clear"
Enter
Sleep 1500ms

Type "kdn remove --force demo"
Enter
Wait+Screen@25000ms /Container removed/
Sleep 1000ms

Type "kdn list"
Enter
Wait+Screen@5000ms /No workspaces registered/
Sleep 5000ms
