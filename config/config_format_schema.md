# Schema for bdaudiodump disc config

Type: `array`

<i id="">path: #</i>

&#36;schema: [http://json-schema.org/draft-07/schema#](http://json-schema.org/draft-07/schema#)

 - **_Items_**
 - Type: `object`
 - <i id="/items">path: #/items</i>
 - **_Properties_**
	 - <b id="#/items/properties/disc_volume_key_sha1">disc_volume_key_sha1</b> `required`
		 - Type: `string`
		 - <i id="/items/properties/disc_volume_key_sha1">path: #/items/properties/disc_volume_key_sha1</i>
	 - <b id="#/items/properties/disc_title">disc_title</b> `required`
		 - Type: `string`
		 - <i id="/items/properties/disc_title">path: #/items/properties/disc_title</i>
	 - <b id="#/items/properties/makemkv_prefix">makemkv_prefix</b> `required`
		 - Type: `string`
		 - <i id="/items/properties/makemkv_prefix">path: #/items/properties/makemkv_prefix</i>
	 - <b id="#/items/properties/album_artist">album_artist</b> `required`
		 - Type: `string`
		 - <i id="/items/properties/album_artist">path: #/items/properties/album_artist</i>
	 - <b id="#/items/properties/genre">genre</b> `required`
		 - Type: `string`
		 - <i id="/items/properties/genre">path: #/items/properties/genre</i>
	 - <b id="#/items/properties/release_date">release_date</b> `required`
		 - Type: `string`
		 - <i id="/items/properties/release_date">path: #/items/properties/release_date</i>
	 - <b id="#/items/properties/disc_number">disc_number</b> `required`
		 - Type: `number`
		 - <i id="/items/properties/disc_number">path: #/items/properties/disc_number</i>
	 - <b id="#/items/properties/total_discs">total_discs</b> `required`
		 - Type: `number`
		 - <i id="/items/properties/total_discs">path: #/items/properties/total_discs</i>
	 - <b id="#/items/properties/total_tracks">total_tracks</b> `required`
		 - Type: `number`
		 - <i id="/items/properties/total_tracks">path: #/items/properties/total_tracks</i>
	 - <b id="#/items/properties/cover_container_relative_path">cover_container_relative_path</b>
		 - Type: `string`
		 - <i id="/items/properties/cover_container_relative_path">path: #/items/properties/cover_container_relative_path</i>
	 - <b id="#/items/properties/cover_relative_path">cover_relative_path</b>
		 - Type: `string`
		 - <i id="/items/properties/cover_relative_path">path: #/items/properties/cover_relative_path</i>
	 - <b id="#/items/properties/cover_url">cover_url</b>
		 - Type: `string`
		 - <i id="/items/properties/cover_url">path: #/items/properties/cover_url</i>
	 - <b id="#/items/properties/cover_type">cover_type</b> `required`
		 - Type: `string`
		 - <i id="/items/properties/cover_type">path: #/items/properties/cover_type</i>
	 - <b id="#/items/properties/tracks">tracks</b> `required`
		 - Type: `array`
		 - <i id="/items/properties/tracks">path: #/items/properties/tracks</i>
			 - **_Items_**
			 - Type: `object`
			 - <i id="/items/properties/tracks/items">path: #/items/properties/tracks/items</i>
			 - **_Properties_**
				 - <b id="#/items/properties/tracks/items/properties/number">number</b> `required`
					 - Type: `number`
					 - <i id="/items/properties/tracks/items/properties/number">path: #/items/properties/tracks/items/properties/number</i>
				 - <b id="#/items/properties/tracks/items/properties/title_number">title_number</b> `required`
					 - Type: `string`
					 - <i id="/items/properties/tracks/items/properties/title_number">path: #/items/properties/tracks/items/properties/title_number</i>
				 - <b id="#/items/properties/tracks/items/properties/chapter_numbers">chapter_numbers</b> `required`
					 - Type: `array`
					 - <i id="/items/properties/tracks/items/properties/chapter_numbers">path: #/items/properties/tracks/items/properties/chapter_numbers</i>
						 - **_Items_**
						 - Type: `number`
						 - <i id="/items/properties/tracks/items/properties/chapter_numbers/items">path: #/items/properties/tracks/items/properties/chapter_numbers/items</i>
				 - <b id="#/items/properties/tracks/items/properties/track_title">track_title</b> `required`
					 - Type: `string`
					 - <i id="/items/properties/tracks/items/properties/track_title">path: #/items/properties/tracks/items/properties/track_title</i>
				 - <b id="#/items/properties/tracks/items/properties/artists">artists</b>
					 - Type: `array`
					 - <i id="/items/properties/tracks/items/properties/artists">path: #/items/properties/tracks/items/properties/artists</i>
						 - **_Items_**
						 - Type: `string`
						 - <i id="/items/properties/tracks/items/properties/artists/items">path: #/items/properties/tracks/items/properties/artists/items</i>
