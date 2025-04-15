#!/bin/bash

JSON_NAME=JYK0bzSTZPConDbzq1GL
POST_PATH=LwSatAQuhZR5y9aDE3dIDATQLKH2/JYK0bzSTZPConDbzq1GL

function prepare_folders () {
    for folder in {1..8}; do
        if [ -d "$folder" ]; then
            rm -fR "${folder:?}/"
            mkdir -p "$folder/$POST_PATH"
        fi
    done
}

function write_json () {
    PART_NUMBER=$1
    FILENAME=$PART_NUMBER/$JSON_NAME

    cat header.json >"$FILENAME"

    for ((part=1; part < PART_NUMBER; part++)); do
        cat "part-$part.json" >>"$FILENAME"
        echo "," >>"$FILENAME"
    done

    cat "part-$PART_NUMBER.json" >>"$FILENAME"

    cat footer.json >>"$FILENAME"
}

function create_images () {
    SOURCE_IMAGE=$1
    DESTINATION_FOLDER=$2

    if [ ! -d "$DESTINATION_FOLDER" ]; then
        mkdir -p "$DESTINATION_FOLDER"
    fi

	convert "$SOURCE_IMAGE" -resize "100x100" "$DESTINATION_FOLDER/small.webp"
	convert "$SOURCE_IMAGE" -resize "500x500" "$DESTINATION_FOLDER/medium.webp"
	convert "$SOURCE_IMAGE" -resize "1000x1000" "$DESTINATION_FOLDER/large.webp"
	convert "$SOURCE_IMAGE" "$DESTINATION_FOLDER/full.webp"
}

# Clear folders and create new ones
prepare_folders

# Create .json files
for part_count in {1..8}; do
    write_json "$part_count"
done

# Create images
create_images bhproxy-1.png 1/$POST_PATH/17976527732702219

create_images bhproxy-2.png 2/$POST_PATH/18061314082790784
cp -r 1/$POST_PATH/* 2/$POST_PATH/

create_images bhproxy-3.png 3/$POST_PATH/18481920520033371
cp -r 2/$POST_PATH/* 3/$POST_PATH/

create_images bhproxy-4.png 4/$POST_PATH/18057176492310232
cp -r 3/$POST_PATH/* 4/$POST_PATH/

create_images bhproxy-5.png 5/$POST_PATH/18288272260174106
cp -r 4/$POST_PATH/* 5/$POST_PATH/

create_images bhproxy-6.png 6/$POST_PATH/18104250796472420
cp -r 5/$POST_PATH/* 6/$POST_PATH/

create_images bhproxy-7.png 7/$POST_PATH/18072041227687192
cp -r 6/$POST_PATH/* 7/$POST_PATH/

create_images bhproxy-8.png 8/$POST_PATH/18045638831111545
cp -r 7/$POST_PATH/* 8/$POST_PATH/
