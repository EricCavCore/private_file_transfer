Live version hosted at `https://sft.ericc.ninja`

Upload files with `curl -F file=@[file.path] https://sft.ericc.ninja`. This will return you a `"https://sft.ericc.ninja?id=[id]&key=[key]"`

Download files with `curl https://sft.ericc.ninja?id=[id]&key=[key] > [file.path]`

Delete files with `curl -X "DELETE" https://sft.ericc.ninja?id=[id]`
