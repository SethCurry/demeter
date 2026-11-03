---
title: "Photos and Timelapse"
summary: "How camera frames travel from an ESP32 to disk, how the Photos page plays them back, and what to check when images do not appear."
date: 2023-09-07T16:13:18+02:00
lastmod: 2026-11-03T00:00:00+00:00
draft: false
weight: 10
toc: true
---

Demeter receives still frames from ESP32 camera devices and stores them as
JPEG files on the server. The **Photos** page lists the cameras that have sent
frames and can play their frames back as an MJPEG stream.

<figure>
  <img src="/images/sample.jpg" alt="Photos page with timelapse cards" width="1200" height="675" loading="lazy">
  <figcaption>Each card represents one camera directory. Replace this placeholder with a screenshot.</figcaption>
</figure>

## How capture works

The stock ESP32 firmware registers a timelapse sensor alongside the scalar
sensors. The timelapse sensor:

1. Initialises the camera and, if the module supports it, autofocus.
2. Captures a frame at most every **5 minutes**. This interval is fixed in
   firmware (`TIMELAPSE_SENSOR_INTERVAL_MS`) and is independent of
   `SENSOR_INTERVAL_SECONDS`, because a 5 MP JPEG takes seconds to upload.
3. Reads the station MAC address and formats it as 12 lowercase hex
   characters, for example `a1b2c3d4e5f6`.
4. Streams the frame over the same WebSocket used for readings.

The stock firmware captures 5 MP JPEGs at quality 12 with a single frame
buffer, then uploads the image in 512-byte chunks with a 30-second timeout. It
retries a capture up to 10 times before giving up. A slower or less reliable
link may therefore delay or skip a frame.

There is no working capture button in the web panel. Cameras decide their own
schedule, and the server stores whatever arrives.

## The binary frame format

Frames are WebSocket **binary** messages, not text readings:

```text
+----------------------+-------------------------------+
| 12-byte MAC (ASCII)  | JPEG bytes                    |
+----------------------+-------------------------------+
```

The firmware sends the 12-byte MAC as the first fragment, then the JPEG in
512-byte continuation frames, then a final frame. The server reassembles the
WebSocket message, treats the first 12 bytes as the MAC string, and treats the
remainder as JPEG data.

A sensor connection handles **one image and then returns**. The server stops
reading that connection but does not explicitly close it. The stock firmware
keeps using that same connection for later scalar readings and does not
deliberately reconnect after the image. Those later messages can be sent but
never read. If scalar readings stop arriving after the first frame, reconnect
the device or run the camera on its own connection.

> **Ordering note:** send all text readings for a connection before the binary
> image. Once the server receives a binary frame it saves the image and stops
> reading that connection.

The server reads the first 12 bytes as the MAC without checking that the
message is at least 12 bytes long. A shorter binary frame can panic the
handler and leave the connection unserviced; always send a complete prefix and
JPEG.

## Where frames are stored

The server writes each frame to:

```text
/srv/demeter/timelapse/<mac>/<RFC3339 timestamp>.jpeg
```

For example:

```text
/srv/demeter/timelapse/a1b2c3d4e5f6/2026-11-03T14:05:12Z.jpeg
```

The per-device directory is created recursively when the first frame arrives,
so the process needs permission to create and write under `/srv/demeter`.
Creating `/srv/demeter/timelapse` yourself and giving the server write access
is the most predictable setup. Failures creating the directory or opening the
file are not logged, and a failed write is ignored. The server can even log
`saved timelapse image` after a failed write, because it does not check the
write result. Verify that files actually exist on disk when diagnosing a
missing image.

### Docker warning

The repository's Compose stack mounts only the SQLite database volume. It does
**not** mount `/srv/demeter/timelapse`, so images stored in a container are
lost when that container is recreated. For a long-lived installation, add a
volume or bind mount for `/srv/demeter/timelapse`.

## The Photos page

The page calls `GET /api/timelapse`, which lists entries under the timelapse
root. Each entry becomes a card showing its name — normally the camera MAC
address.

The play button links to:

```text
/api/timelapse/<mac>/timelapse.mjpeg
```

That endpoint:

1. Reads every `.jpeg` file in the MAC directory in filename order. Because
   filenames are RFC3339 timestamps, this is chronological order.
2. Vertically flips each image with `imaging.FlipV`.
3. Re-encodes it as JPEG.
4. Pushes it into an MJPEG stream, waiting about 200 ms between frames.
5. Closes the stream after the last frame.

The stream is generated on demand every time you open it. It is not a stored
video file. Opening the link in a browser or media player that supports MJPEG
plays the sequence; a player that only accepts MP4 will not.

The **New Timelapse** button and the preview panel are present in the current
UI but are not wired to an action. Use the play link on a card.

## Filename ordering and time zones

Filenames come from the server's `time.Now()` formatted as RFC3339. They sort
lexicographically in chronological order as long as the server's UTC offset
does not change mid-recording. If you move the server between time zones,
frames from before and after the move may interleave oddly. Running the server
with `TZ=UTC` avoids this.

## Retention

There is no automatic retention or cleanup for timelapse images. Frames
accumulate until you delete them. A simple cron job can prune old files:

```sh
# Delete frames older than 30 days.
find /srv/demeter/timelapse -type f -name '*.jpeg' -mtime +30 -delete
```

Stop the server or accept that a frame may be written while cleanup runs. Do
not delete the MAC directories themselves unless you also want them to
disappear from the Photos page.

## Troubleshooting

**No cards on the Photos page.** The server has not received a binary frame
yet, or `/srv/demeter/timelapse` is empty. If the root directory itself is
missing or unreadable, the list handler can return a `500` (often with an
empty `[]` body) because it continues after the directory read fails. Check
the server log for
`received timelapse image message` and `saved timelapse image`. If you see
`failed to take picture`, the camera did not return a frame. A `saved` log
alone does not guarantee that the file was written successfully.

**Card exists but playback is blank or errors.** Check that the MAC directory
contains `.jpeg` files and that the server can read them. Corrupt frames are
skipped with a log message. The playback route returns an error if it cannot
read the directory.

**Images disappear after a restart.** In Docker, the timelapse directory is
not on a persistent volume. Add one.

**Images are upside down.** The server flips frames vertically before
streaming. If your camera is mounted the other way, the stream will appear
inverted; there is currently no per-camera orientation setting.

See [Troubleshooting](../troubleshooting) for server and sensor checks.
