# hypersonic

[frontend](https://github.com/dragsbruh/hypersonic-ui)

## todo

- [ ] database queries

  - [ ] albums
    - [x] create album
    - [x] get album by hash (includes album artists)
    - [x] get albums by artist
    - [ ] search albums by name
    - [ ] paginate (sort by params + direction)
    - [x] delete album

  - [ ] tracks
    - [x] create track
    - [x] get track by hash (includes artists and albums, does not include album artists)
    - [x] get tracks by album
    - [x] get tracks by artist
    - [ ] get tracks by genre
    - [ ] search tracks by name
    - [x] paginate (sort by params + direction)
    - [x] delete track

  - [ ] artists
    - [x] create artist
    - [x] get artist by hash
    - [ ] search artists
    - [x] delete artist

### meta

- [ ] share duplicated code
- [ ] handle row not found
- [ ] this is getting messy, need to rewrite sql with json agg
