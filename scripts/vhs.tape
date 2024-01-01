# VHS documentation
#
# Output:
#   Output <path>.gif               Create a GIF output at the given <path>
#   Output <path>.mp4               Create an MP4 output at the given <path>
#   Output <path>.webm              Create a WebM output at the given <path>
#
# Settings:
#   Set FontSize <number>           Set the font size of the terminal
#   Set FontFamily <string>         Set the font family of the terminal
#   Set Height <number>             Set the height of the terminal
#   Set Width <number>              Set the width of the terminal
#   Set LetterSpacing <float>       Set the font letter spacing (tracking)
#   Set LineHeight <float>          Set the font line height
#   Set Theme <string>              Set the theme of the terminal (JSON)
#   Set Padding <number>            Set the padding of the terminal
#   Set Framerate <number>          Set the framerate of the recording
#   Set PlaybackSpeed <float>       Set the playback speed of the recording
#
# Sleep:
#   Sleep <time>                    Sleep for a set amount of <time> in seconds
#
# Type:
#   Type[@<time>] "<characters>"    Type <characters> into the terminal with a
#                                   <time> delay between each character
#
# Keys:
#   Backspace[@<time>] [number]     Press the Backspace key
#   Down[@<time>] [number]          Press the Down key
#   Enter[@<time>] [number]         Press the Enter key
#   Space[@<time>] [number]         Press the Space key
#   Tab[@<time>] [number]           Press the Tab key
#   Left[@<time>] [number]          Press the Left Arrow key
#   Right[@<time>] [number]         Press the Right Arrow key
#   Up[@<time>] [number]            Press the Up Arrow key
#   Down[@<time>] [number]          Press the Down Arrow key
#   Ctrl+<key>                      Press the Control key + <key> (e.g. Ctrl+C)
#
# Display:
#   Hide                            Hide the subsequent commands from the output
#   Show                            Show the subsequent commands in the output

Output "../img/wander.gif"

Set Padding 10
Set FontSize 13
Set Width 1100
Set Height 450
Set Framerate 15

Hide
Type "wander"
Enter
Sleep 500ms
Show

# All Jobs
Sleep 1s
Type@500ms "jj"
Type@500ms "k"
Sleep 101ms
Screenshot ../img/screenshots/All_Jobs.png
Sleep 1s

# All Events
Type "V"
Sleep 100ms
Screenshot ../img/screenshots/Global_Events.png
Sleep 2s
Type@300ms "j"
Enter
Sleep 1s
Type@500ms "dd"

# Back to All Jobs
Escape@1s 2
Sleep 1s

# All Tasks
Type "A"
Sleep 100ms
Screenshot ../img/screenshots/All_Tasks.png
Sleep 1s
Type@500ms "jj"
Type@500ms "k"
Sleep 1s

# Exec for Task
Type "e"
Sleep 1s
Backspace 10
Type@200ms "bash"
Sleep 500ms
Enter
Type@150ms "ls -la"
Sleep 300ms
Enter
Sleep 100ms
Screenshot ../img/screenshots/Exec.png
Sleep 2s
Ctrl+d
Sleep 2s
Escape@300ms 2
Sleep 500ms

# Back to All Jobs
Type "J"
Sleep 1s

# Tasks for Job
Enter
Sleep 100ms
Screenshot ../img/screenshots/Tasks_for_Job.png
Sleep 2s

# Logs for Task, Filtering
Enter
Sleep 100ms
Screenshot ../img/screenshots/Task_Logs.png
Sleep 2s
Type "/"
Sleep 1s
Type@200ms "INFO"
Sleep 0.5s
Enter
Type@0.5s "nnnN"
Type@300ms "jjkk"
Enter
Sleep 1s
Escape@200ms 2

# Save Logs
Ctrl+S
Type@100ms "/tmp/my_logs.txt"
Sleep 500ms
Enter
Sleep 500ms
Screenshot ../img/screenshots/Save_Any_View_to_Local_File.png
Sleep 2.5s

# Back to Tasks for Job
Escape
Sleep 1s

# Stats for Allocation
Type "s"
Sleep 100ms
Screenshot ../img/screenshots/Allocation_Statistics.png
Sleep 2s

# Exit
Hide
Ctrl+c
Sleep 1s
