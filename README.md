
Why not just use mv, cp, and rm??

Here's a scenerio you're downloading tons of stuff into your downloads folder, this will be useful + i was just building random stuff

Note: A config file must be available

Here's a template
```json
{
    "file": [
        {
        "extensions": ["mp3"],
        "watch": "<path to directory>",
        "destination": "<path to mp3 directory>"
        },
        {
        "extensions": ["mp4"],
        "watch": "<path to directory>",
        "destination": "<path to mp3 directory>"
        },
    ],
}
```