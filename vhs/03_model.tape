Output demos/03_model.gif

Set Shell "bash"
Set FontSize 20
Set Width 1600
Set Height 600
Set Padding 16
Set Framerate 15
Set TypingSpeed 30ms

Type "export KDN_DEFAULT_RUNTIME=podman"
Enter
Sleep 500ms

Type "export KDN_DEFAULT_AGENT=claude"
Enter
Sleep 500ms

Type "export KDN_INIT_AUTO_START=1"
Enter
Sleep 500ms

Type "kdn init --model opus"
Enter
Wait+Screen@5000ms /Container started/
Sleep 2000ms

Type "kdn terminal demo"
Enter
Sleep 5000ms

Hide

Type "/exit"
Enter
Sleep 2000ms

Type "kdn remove --force demo"
Enter
Wait+Screen@25000ms /Container removed/
