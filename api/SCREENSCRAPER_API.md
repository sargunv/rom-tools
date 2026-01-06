# ScreenScraper API Documentation

> [!WARNING]
>
> This document is a machine translation of the [original documentation](https://www.screenscraper.fr/webapi2.php) on 2026-01-05. There may be out-of-date information or errors.

## API Overview

**Warning**: Version 2 is in beta test mode. Modifications may be made to this API version at any time without notice.

Our API allows you to obtain all ScreenScraper data and media to integrate into your applications: front-ends, utilities. All our requests return the requested information in XML, JSON, or INI format.

### Who can use the API?

The ScreenScraper API can only be integrated into applications that are entirely free and distributed, or, otherwise, with prior authorization and conditions set by the ScreenScraper team. Any violation of this rule may result in account termination or legal action.

If you are a developer and want to integrate our API, [contact us via the forum](https://www.screenscraper.fr/forumsujets.php?frub=12&numpage=0) to present your software and obtain your credentials and password to provide to the API to validate your usage rights.

### How to make requests to our API?

Information and/or media requests to the ScreenScraper API are made by calling URL requests of type GET, and an XML or JSON document is returned.

**Example Game Search:**

```
https://api.screenscraper.fr/api2/jeuInfos.php?devid=xxx&devpassword=yyy&softname=zzz&ssid=test&sspassword=test&output=xml&crc=50ABC90A&systemeid=1&romtype=rom&romnom=Sonic%20The%20Hedgehog%202%20(World).zip&romtaille=749652
```

## How many simultaneous requests to the API are allowed?

Depending on the user's contribution level (to the database or financial), they are assigned "Thread" openings.

### What is a "Thread"?

Some scraping software, like "Universal XML Scraper", allows you to save scraping time by not processing ROMs one by one, but simultaneously in parallel. The number of simultaneous scrapes is what we call a "Thread". So with 4 Threads, you scrape your ROMs 4 by 4, which represents a huge time savings. However, this requires much more from your computer and bandwidth (so it needs to have the necessary resources), but especially, it requires much more from the database and the server hosting it. This is why the "reward" system was put in place. To thank you for contributing to the ScreenScraper database, we offer you the ability to draw from its resources more efficiently.

### How to earn "Threads"?

There are 2 simple methods:

- Participate in the database by proposing new information or new media
- Participate financially in hosting the database via Tipee or Patreon

### How many "Threads" can I earn?

Consult the [FAQ](https://www.screenscraper.fr/faq.php) for more information.

## How many requests to the API are allowed per minute and per day?

Since approximately mid-2019, a "Quota" system has been integrated into the API to avoid saturating our servers.

The principle is to limit API access rights for each user **per minute** and **per day** based on their level and financial participation.

Consult the [FAQ](https://www.screenscraper.fr/faq.php) for more information.

Quota information is returned in user data so that it can be easily integrated and managed directly by the scraping software.

This quota management by software is now mandatory to avoid saturating our servers unnecessarily.

## Error Responses

API requests return HTTP error numbers in case of problems. Here they are:

| Error | Description | Cause |
| --- | --- | --- |
| 400 | Problem with URL | The API call URL contains no information |
| 400 | Missing required fields in URL | One of the minimum required fields is missing in the API call URL |
| 400 | Error in ROM filename: contains a path | The ROM filename sent is of type "/mnt/sda1/batocera/roms/..." |
| 400 | CRC, MD5 or SHA1 field error | The CRC, MD5 or SHA1 field is not correctly formatted |
| 400 | Problem in ROM filename | The ROM filename is not compliant |
| 401 | API closed for non-members or inactive members | Server is saturated (CPU usage >60%) |
| 403 | Login error: Check your developer credentials! | Incorrect developer credentials |
| 404 | Error: Game not found! / Error: Rom/Iso/Folder not found! | Unable to find a match for the requested ROM |
| 423 | API completely closed | Server has serious problems |
| 426 | The scraping software used has been blacklisted (non-compliant / obsolete version) | Need to change software version |
| 429 | The number of threads allowed for the member is reached | Need to reduce request speed |
| 429 | The number of threads per minute allowed for the member is reached | Need to reduce request speed |
| 429 | The maximum threads allowed to leecher users is already used | Need to reduce request speed |
| 429 | The maximum threads is already used | Need to reduce request speed |
| 430 | Your scraping quota is exceeded for today! | Member has scraped more than x (see FAQ) ROMs in the day |
| 431 | Clean up your ROM files and come back tomorrow! | Member has scraped more than x (see FAQ) ROMs not recognized by ScreenScraper |

## API Endpoints List

- **ssinfraInfos.php**: ScreenScraper infrastructure information
- **ssuserInfos.php**: ScreenScraper user information
- **userlevelsListe.php**: List of ScreenScraper user levels
- **nbJoueursListe.php**: List of player counts
- **supportTypesListe.php**: List of support types
- **romTypesListe.php**: List of ROM types
- **regionsListe.php**: List of regions
- **languesListe.php**: List of languages
- **genresListe.php**: List of genres
- **famillesListe.php**: List of families
- **classificationsListe.php**: List of classifications (Game Rating)
- **mediasSystemeListe.php**: List of media for systems
- **mediasJeuListe.php**: List of media for games
- **infosJeuListe.php**: List of info for games
- **infosRomListe.php**: List of info for ROMs
- **mediaGroup.php**: Download game group image media
- **mediaCompagnie.php**: Download game company image media
- **systemesListe.php**: List of systems / system information / system media information
- **mediaSysteme.php**: Download system image media
- **mediaVideoSysteme.php**: Download system video media
- **jeuRecherche.php**: Search for a game by name (returns a table of games (limited to 30 games) sorted by probability)
- **jeuInfos.php**: Game information / Game media
- **mediaJeu.php**: Download game image media
- **mediaVideoJeu.php**: Download game video media
- **mediaManuelJeu.php**: Download game manuals
- **botNote.php**: System for automated sending of game ratings by ScreenScraper members
- **botProposition.php**: System for automated sending of info or media proposals to ScreenScraper

---

## ssinfraInfos.php: ScreenScraper Infrastructure Information

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **output**: xml (default), json

### Returned Elements

**Item: serveurs** (ScreenScraper server information)

- **cpu1**: CPU usage % of server 1 (average of last 5 minutes)
- **cpu2**: CPU usage % of server 2 (average of last 5 minutes)
- **cpu3**: CPU usage % of server 3 (average of last 5 minutes)
- **threadsmin**: Number of API accesses since the last minute
- **nbscrapeurs**: Number of scrapers using the API since the last minute
- **apiacces**: Number of API accesses in the current day (GMT+1)

**Status**

- **closefornomember**: API closed for anonymous users (not registered or not identified) (0: open / 1: closed)
- **closeforleecher**: API closed for non-participating members (no validated proposals) (0: open / 1: closed)

**Quota**

- **maxthreadfornonmember**: Maximum number of threads opened for anonymous users (not registered or not identified) simultaneously by the API
- **threadfornonmember**: Current number of threads opened by anonymous users (not registered or not identified) simultaneously by the API
- **maxthreadformember**: Maximum number of threads opened for members simultaneously by the API
- **threadformember**: Current number of threads opened by members simultaneously by the API

**Example call:**

```
https://api.screenscraper.fr/api2/ssinfraInfos.php?devid=xxx&devpassword=yyy&softname=zzz&output=xml
```

---

## ssuserInfos.php: ScreenScraper User Information

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **output**: xml (default), json
- **ssid**: ScreenScraper user identifier
- **sspassword**: ScreenScraper user password

### Returned Elements

**Item: ssuser** (ScreenScraper user information)

- **id**: User's username on ScreenScraper
- **numid**: Numeric identifier of the user on ScreenScraper
- **niveau**: User's level on ScreenScraper
- **contribution**: Financial contribution level on ScreenScraper (2 = 1 Additional Thread / 3 and + = 5 Additional Threads)
- **uploadsysteme**: Counter of valid contributions (system media) proposed by the user
- **uploadinfos**: Counter of valid contributions (text info) proposed by the user
- **romasso**: Counter of valid contributions (ROM association) proposed by the user
- **uploadmedia**: Counter of valid contributions (game media) proposed by the user
- **propositionok**: Number of user proposals validated by a moderator
- **propositionko**: Number of user proposals rejected by a moderator
- **quotarefu**: Percentage of proposal rejection by the user

**Threads**

- **maxthreads**: Number of threads allowed for the user (also indicated for non-registered users)
- **maxdownloadspeed**: Download speed (in KB/s) allowed for the user (also indicated for non-registered users)

**Quotas**

- **requeststoday**: Total number of API calls during the current day
- **requestskotoday**: Number of API calls with negative return (ROM/game not found) during the current day
- **maxrequestspermin**: Maximum number of API calls allowed per minute for the user (see FAQ)
- **maxrequestsperday**: Maximum number of API calls allowed per day for the user (see FAQ)
- **maxrequestskoperday**: Maximum number of API calls with negative return (ROM/game not found) allowed per day for the user (see FAQ)
- **visites**: Number of user visits to ScreenScraper
- **datedernierevisite**: Date of the user's last visit to ScreenScraper (format: yyyy-mm-dd hh:mm:ss)
- **favregion**: Favorite region of user visits to ScreenScraper (france, europe, usa, japan)

**Example call:**

```
https://api.screenscraper.fr/api2/ssuserInfos.php?devid=xxx&devpassword=yyy&softname=zzz&output=xml&ssid=test&sspassword=test
```

---

## userlevelsListe.php: List of ScreenScraper User Levels

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **output**: xml (default), json
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password

### Returned Elements

**Items: userlevels**

- **userlevel** (xml) / _id_ (json):
  - **id**: Numeric identifier of the level
  - **nom_fr**: Name of the level in French

**Example call:**

```
https://api.screenscraper.fr/api2/userlevelsListe.php?devid=xxx&devpassword=yyy&softname=zzz&output=xml&ssid=test&sspassword=test
```

---

## nbJoueursListe.php: List of Player Counts

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **output**: xml (default), json
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password

### Returned Elements

**Items: nbjoueur**

- **id**: Numeric identifier of the player count
- **nom**: Designation of the player count
- **parent**: Numeric identifier of the parent player count (0 if no parent)

**Example call:**

```
https://api.screenscraper.fr/api2/nbJoueursListe.php?devid=xxx&devpassword=yyy&softname=zzz&output=XML&ssid=test&sspassword=test
```

---

## supportTypesListe.php: List of Support Types

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **output**: xml (default), json
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password

### Returned Elements

**Items: supporttypes**

- **nom**: Designation of the support(s)

**Example call:**

```
https://api.screenscraper.fr/api2/supportTypesListe.php?devid=xxx&devpassword=yyy&softname=zzz&output=XML&ssid=test&sspassword=test
```

---

## romTypesListe.php: List of ROM Types

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **output**: xml (default), json
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password

### Returned Elements

**Items: romtypes**

- **nom**: Designation of the ROM type(s)

**Example call:**

```
https://api.screenscraper.fr/api2/romTypesListe.php?devid=xxx&devpassword=yyy&softname=zzz&output=xml&ssid=test&sspassword=test
```

---

## genresListe.php: List of Genres

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **output**: xml (default), json
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password

### Returned Elements

**Items: genres**

- **genre** (xml) / _id_ (json):
  - **id**: Numeric identifier of the genre
  - **nom_de**: Name of the genre in German
  - **nom_en**: Name of the genre in English
  - **nom_es**: Name of the genre in Spanish
  - **nom_fr**: Name of the genre in French
  - **nom_it**: Name of the genre in Italian
  - **nom_pt**: Name of the genre in Portuguese
  - **parent**: ID of the parent genre (0 if main genre)
  - **medias**:
    - **media_pictomonochrome**: Media download URL: Monochrome Pictogram
    - **media_pictomonochromesvg**: Media download URL: Monochrome Vector Pictogram
    - **media_pictocouleur**: Media download URL: Color Pictogram
    - **media_pictocouleursvg**: Media download URL: Color Vector Pictogram
    - **media_background**: Media download URL: Background

**Example call:**

```
https://api.screenscraper.fr/api2/genresListe.php?devid=xxx&devpassword=yyy&softname=zzz&output=xml&ssid=test&sspassword=test
```

---

## famillesListe.php: List of Families

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **output**: xml (default), json
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password

### Returned Elements

**Items: familles**

- **famille** (xml) / _id_ (json):
  - **id**: Numeric identifier of the family
  - **nom**: Name of the family
  - **medias**:
    - **media_pictomonochrome**: Media download URL: Monochrome Pictogram
    - **media_pictomonochromesvg**: Media download URL: Monochrome Vector Pictogram
    - **media_pictocouleur**: Media download URL: Color Pictogram
    - **media_pictocouleursvg**: Media download URL: Color Vector Pictogram
    - **media_background**: Media download URL: Background
    - **media_figurine**: Media download URL: Figurine

**Example call:**

```
https://api.screenscraper.fr/api2/famillesListe.php?devid=xxx&devpassword=yyy&softname=zzz&output=xml&ssid=test&sspassword=test
```

---

## regionsListe.php: List of Regions

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **output**: xml (default), json
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password

### Returned Elements

**Items: regions**

- **region** (xml) / _id_ (json):
  - **id**: Numeric identifier of the region
  - **nomcourt**: Short name of the region
  - **nom_de**: Name of the region in German
  - **nom_en**: Name of the region in English
  - **nom_es**: Name of the region in Spanish
  - **nom_fr**: Name of the region in French
  - **nom_it**: Name of the region in Italian
  - **nom_pt**: Name of the region in Portuguese
  - **parent**: ID of the parent region (0 if main region)
  - **medias**:
    - **media_pictomonochrome**: Media download URL: Monochrome Pictogram
    - **media_pictomonochromesvg**: Media download URL: Monochrome Vector Pictogram
    - **media_pictocouleur**: Media download URL: Color Pictogram
    - **media_pictocouleursvg**: Media download URL: Color Vector Pictogram
    - **media_background**: Media download URL: Background

**Example call:**

```
https://api.screenscraper.fr/api2/regionsListe.php?devid=xxx&devpassword=yyy&softname=zzz&output=xml&ssid=test&sspassword=test
```

---

## languesListe.php: List of Languages

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **output**: xml (default), json
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password

### Returned Elements

**Items: langues**

- **langue** (xml) / _id_ (json):
  - **id**: Numeric identifier of the language
  - **nomcourt**: Short name of the language
  - **nom_de**: Name of the language in German
  - **nom_en**: Name of the language in English
  - **nom_es**: Name of the language in Spanish
  - **nom_fr**: Name of the language in French
  - **nom_it**: Name of the language in Italian
  - **nom_pt**: Name of the language in Portuguese
  - **parent**: ID of the parent language (0 if main language)
  - **medias**:
    - **media_pictomonochrome**: Media download URL: Monochrome Pictogram
    - **media_pictomonochromesvg**: Media download URL: Monochrome Vector Pictogram
    - **media_pictocouleur**: Media download URL: Color Pictogram
    - **media_pictocouleursvg**: Media download URL: Color Vector Pictogram
    - **media_background**: Media download URL: Background

**Example call:**

```
https://api.screenscraper.fr/api2/languesListe.php?devid=xxx&devpassword=yyy&softname=zzz&output=xml&ssid=test&sspassword=test
```

---

## classificationsListe.php: List of Classifications (Game Rating)

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **output**: xml (default), json
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password

### Returned Elements

**Items: classifications**

- **langue** (xml) / _id_ (json):
  - **id**: Numeric identifier of the classification
  - **nomcourt**: Short name of the classification
  - **nom_de**: Name of the classification in German (if exists)
  - **nom_en**: Name of the classification in English (if exists)
  - **nom_es**: Name of the classification in Spanish (if exists)
  - **nom_fr**: Name of the classification in French (if exists)
  - **nom_it**: Name of the classification in Italian (if exists)
  - **nom_pt**: Name of the classification in Portuguese (if exists)
  - **parent**: ID of the parent classification (0 if main classification)
  - **medias**:
    - **media_pictomonochrome**: Media download URL: Monochrome Pictogram
    - **media_pictomonochromesvg**: Media download URL: Monochrome Vector Pictogram
    - **media_pictocouleur**: Media download URL: Color Pictogram
    - **media_pictocouleursvg**: Media download URL: Color Vector Pictogram
    - **media_background**: Media download URL: Background

**Example call:**

```
https://api.screenscraper.fr/api2/classificationsListe.php?devid=xxx&devpassword=yyy&softname=zzz&output=xml&ssid=test&sspassword=test
```

---

## mediasSystemeListe.php: List of Media for Systems

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **output**: xml (default), json
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password

### Returned Elements

**Items: medias**

- **media** (xml) / _id_ (json):
  - **id**: Numeric identifier of the media
  - **nomcourt**: Short name of the media
  - **nom**: Long name of the media
  - **categorie**: Category of the media
  - **plateformtypes**: List of system types where the media is present (system type ID separated by |, if empty = all system types)
  - **plateforms**: List of systems where the media is present (system ID separated by |, if empty = all systems)
  - **type**: Media type
  - **fileformat**: File format of the media
  - **fileformat2**: Second file format of the media accepted for proposals
  - **autogen**: Auto-generated media (0=no, 1=yes)
  - **multiregions**: Multi-region media (0=no, 1=yes)
  - **multisupports**: Multi-support media (0=no, 1=yes)
  - **multiversions**: Multi-version media (0=no, 1=yes)
  - **extrainfostxt**: Additional information about the media

**Example call:**

```
https://api.screenscraper.fr/api2/mediasSystemeListe.php?devid=xxx&devpassword=yyy&softname=zzz&output=xml&ssid=test&sspassword=test
```

---

## mediasJeuListe.php: List of Media for Games

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **output**: xml (default), json
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password

### Returned Elements

**Items: medias**

- **media** (xml) / _id_ (json):
  - **id**: Numeric identifier of the media
  - **nomcourt**: Short name of the media
  - **nom**: Long name of the media
  - **categorie**: Category of the media
  - **plateformtypes**: List of system types where the media is present (system type ID separated by |, if empty = all system types)
  - **plateforms**: List of systems where the media is present (system ID separated by |, if empty = all systems)
  - **type**: Media type
  - **fileformat**: File format of the media
  - **fileformat2**: Second file format of the media accepted for proposals
  - **autogen**: Auto-generated media (0=no, 1=yes)
  - **multiregions**: Multi-region media (0=no, 1=yes)
  - **multisupports**: Multi-support media (0=no, 1=yes)
  - **multiversions**: Multi-version media (0=no, 1=yes)
  - **extrainfostxt**: Additional information about the media

**Example call:**

```
https://api.screenscraper.fr/api2/mediasJeuListe.php?devid=xxx&devpassword=yyy&softname=zzz&output=xml&ssid=test&sspassword=test
```

---

## infosJeuListe.php: List of Info for Games

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **output**: xml (default), json
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password

### Returned Elements

**Items: infos**

- **info** (xml) / _id_ (json):
  - **id**: Numeric identifier of the info
  - **nomcourt**: Short name of the info
  - **nom**: Long name of the info
  - **categorie**: Category of the info
  - **plateformtypes**: List of system types where the info is present (system type ID separated by |, if empty = all system types)
  - **plateforms**: List of systems where the info is present (system ID separated by |, if empty = all systems)
  - **type**: Info type
  - **autogen**: Auto-generated info (0=no, 1=yes)
  - **multiregions**: Multi-region info (0=no, 1=yes)
  - **multisupports**: Multi-support info (0=no, 1=yes)
  - **multiversions**: Multi-version info (0=no, 1=yes)
  - **multichoix**: Multi-choice info (0=no, 1=yes)

**Example call:**

```
https://api.screenscraper.fr/api2/infosJeuListe.php?devid=xxx&devpassword=yyy&softname=zzz&output=xml&ssid=test&sspassword=test
```

---

## infosRomListe.php: List of Info for ROMs

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **output**: xml (default), json
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password

### Returned Elements

**Items: infos**

- **info** (xml) / _id_ (json):
  - **id**: Numeric identifier of the info
  - **nomcourt**: Short name of the info
  - **nom**: Long name of the info
  - **categorie**: Category of the info
  - **plateformtypes**: List of system types where the info is present (system type ID separated by |, if empty = all system types)
  - **plateforms**: List of systems where the info is present (system ID separated by |, if empty = all systems)
  - **type**: Info type
  - **autogen**: Auto-generated info (0=no, 1=yes)
  - **multiregions**: Multi-region info (0=no, 1=yes)
  - **multisupports**: Multi-support info (0=no, 1=yes)
  - **multiversions**: Multi-version info (0=no, 1=yes)
  - **multichoix**: Multi-choice info (0=no, 1=yes)

**Example call:**

```
https://api.screenscraper.fr/api2/infosRomListe.php?devid=xxx&devpassword=yyy&softname=zzz&output=xml&ssid=test&sspassword=test
```

---

## mediaGroup.php: Download Game Group Image Media

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password
- **crc**: CRC calculation of existing local image
- **md5**: MD5 calculation of existing local image
- **sha1**: SHA1 calculation of existing local image
- **groupid**: Numeric identifier of the group (see genreListe.php, modeListe.php,... / group types = genre, mode, famille, theme, style)
- **media**: Text identifier of the media to return (see genreListe.php, modeListe.php,... / group types = genre, mode, famille, theme, style)
- **mediaformat** (optional): Media format (extension): ex: jpg, png, mp4, zip, mp3, ... (informative: does not return the media in the specified format)

### Output Parameters

- **maxwidth** (optional): Maximum width in pixels of the returned image
- **maxheight** (optional): Maximum height in pixels of the returned image
- **outputformat** (optional): Format (extension) of the returned image: png or jpg

### Returned Element

PNG Image or Text CRCOK or MD5OK or SHA1OK if the crc, md5 or sha1 parameter is identical to the crc, md5 or sha1 calculation of the server image (update optimization) or Text NOMEDIA if the media file was not found

**Example call:**

```
https://api.screenscraper.fr/api2/mediaGroup.php?devid=xxx&devpassword=yyy&softname=zzz&ssid=test&sspassword=test&crc=&md5=&sha1=&groupid=1&media=logo-monochrome
```

---

## mediaCompagnie.php: Download Game Company Image Media

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password
- **crc**: CRC calculation of existing local image
- **md5**: MD5 calculation of existing local image
- **sha1**: SHA1 calculation of existing local image
- **companyid**: Numeric identifier of the company
- **media**: Text identifier of the media to return
- **mediaformat** (optional): Media format (extension): ex: jpg, png, mp4, zip, mp3, ... (informative: does not return the media in the specified format)

### Output Parameters

- **maxwidth** (optional): Maximum width in pixels of the returned image
- **maxheight** (optional): Maximum height in pixels of the returned image
- **outputformat** (optional): Format (extension) of the returned image: png or jpg

### Returned Element

PNG Image or Text CRCOK or MD5OK or SHA1OK if the crc, md5 or sha1 parameter is identical to the crc, md5 or sha1 calculation of the server image (update optimization) or Text NOMEDIA if the media file was not found

**Example call:**

```
https://api.screenscraper.fr/api2/mediaCompagnie.php?devid=xxx&devpassword=yyy&softname=zzz&ssid=test&sspassword=test&crc=&md5=&sha1=&companyid=3&media=logo-monochrome
```

---

## systemesListe.php: List of Systems / System Information / System Media Information

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **output**: xml (default), json
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password

### Returned Elements

**Item: ssuser** (ScreenScraper user information)

- **id**: User's username on ScreenScraper
- **niveau**: User's level on ScreenScraper
- **contribution**: Financial contribution level on ScreenScraper (2 = 1 Additional Thread / 3 and + = 5 Additional Threads)
- **uploadsysteme**: Counter of valid contributions (system media) proposed by the user
- **uploadinfos**: Counter of valid contributions (text info) proposed by the user
- **romasso**: Counter of valid contributions (ROM association) proposed by the user
- **uploadmedia**: Counter of valid contributions (game media) proposed by the user
- **maxthreads**: Number of threads allowed for the user (also indicated for non-registered users)
- **maxdownloadspeed**: Download speed (in KB/s) allowed for the user (also indicated for non-registered users)
- **requeststoday**: Total number of API calls during the current day
- **requestskotoday**: Number of API calls with negative return (ROM/game not found) during the current day
- **maxrequestsperdmin**: Maximum number of API calls allowed per minute for the user (see FAQ)
- **maxrequestsperday**: Maximum number of API calls allowed per day for the user (see FAQ)
- **maxrequestskoperday**: Maximum number of API calls with negative return (ROM/game not found) allowed per day for the user (see FAQ)
- **visites**: Number of user visits to ScreenScraper
- **datedernierevisite**: Date of the user's last visit to ScreenScraper (format: yyyy-mm-dd hh:mm:ss)
- **favregion**: Favorite region of user visits to ScreenScraper (france, europe, usa, japan)

**Items: systeme** (xml) / **systemes** (json)

- **id**: Numeric identifier of the system (to be provided in other API requests)
- **parentid**: Numeric identifier of the parent system
- **noms**:
  - **nom_xx**: System name Region xx (xx = "nomcourt" variable from regionsListe.php API)
  - **nom_recalbox**: System name in Recalbox front-end
  - **nom_retropie**: System name in Retropie front-end
  - **nom_launchbox**: System name in Launchbox front-end
  - **nom_hyperspin**: System name in Hyperspin front-end
  - **noms_commun**: Common names given to the system in general
- **extensions**: Extensions of usable ROM files (all emulators combined)
- **compagnie**: Name of the system production company
- **type**: System type (Arcade, Console, Console Portable, Emulation Arcade, Flipper, Online, Ordinateur, Smartphone)
- **datedebut**: Year of production start
- **datefin**: Year of production end
- **romtype**: ROM type(s) (see romTypesListe request)
- **supporttype**: Original support type(s) of the system (see supportTypesListe request)
- **medias**: Extensive media structure including logos, wheels, photos, videos, fanart, bezels, backgrounds, screen marquees, box art, support images, controllers, illustrations, etc. Each media includes download URLs and CRC32, MD5, SHA1 identifiers for verification.

**Example call:**

```
https://api.screenscraper.fr/api2/systemesListe.php?devid=xxx&devpassword=yyy&softname=zzz&output=XML&ssid=test&sspassword=test
```

---

## mediaSysteme.php: Download System Image Media

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password
- **crc**: CRC calculation of existing local image
- **md5**: MD5 calculation of existing local image
- **sha1**: SHA1 calculation of existing local image
- **systemeid**: Numeric identifier of the system (see systemesListe.php)
- **media**: Text identifier of the media to return (see systemesListe.php)
- **mediaformat** (optional): Media format (extension): ex: jpg, png, mp4, zip, mp3, ... (informative: does not return the media in the specified format)

### Output Parameters

- **maxwidth** (optional): Maximum width in pixels of the returned image
- **maxheight** (optional): Maximum height in pixels of the returned image
- **outputformat** (optional): Format (extension) of the returned image: png or jpg

### Returned Element

PNG Image or Text CRCOK or MD5OK or SHA1OK if the crc, md5 or sha1 parameter is identical to the crc, md5 or sha1 calculation of the server image (update optimization) or Text NOMEDIA if the media file was not found

**Example call:**

```
https://api.screenscraper.fr/api2/mediaSysteme.php?devid=xxx&devpassword=yyy&softname=zzz&ssid=test&sspassword=test&crc=&md5=&sha1=&systemeid=1&media=wheel(wor)
```

---

## mediaVideoSysteme.php: Download System Video Media

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password
- **crc**: CRC calculation of existing local video
- **md5**: MD5 calculation of existing local video
- **sha1**: SHA1 calculation of existing local video
- **systemeid**: Numeric identifier of the system (see systemesListe.php)
- **media**: Text identifier of the media to return (see systemesListe.php)
- **mediaformat**: Media format (extension): ex: jpg, png, mp4, zip, mp3, ... (optional, informative: does not return the media in the specified format)

### Returned Element

MP4 Video or Text CRCOK or MD5OK or SHA1OK if the crc, md5 or sha1 parameter is identical to the crc, md5 or sha1 calculation of the server video (update optimization) or Text NOMEDIA if the media file was not found

**Example call:**

```
https://api.screenscraper.fr/api2/mediaVideoSysteme.php?devid=xxx&devpassword=yyy&softname=zzz&ssid=test&sspassword=test&crc=&md5=&sha1=&systemeid=1&media=video
```

---

## jeuRecherche.php: Search for a Game by Name

Returns a table of games (limited to 30 games) sorted by probability.

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **output**: xml (default), json
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password
- **systemeid** (optional): Numeric identifier of the system (see systemesListe.php)
- **recherche**: Name of the game to search

### Returned Elements

**Item: serveurs** (ScreenScraper server information)

- **serveurcpu1**: % CPU usage on the current main server
- **serveurcpu2**: % CPU usage on the current secondary server
- **threadsmin**: Number of API accesses in the last 60 seconds
- **nbscrapeurs**: Number of API users currently scraping
- **apiacces**: Number of API accesses today (French time)

**Item: ssuser** (ScreenScraper user information)

- **id**: User's username on ScreenScraper
- **niveau**: User's level on ScreenScraper
- **contribution**: Financial contribution level on ScreenScraper (2 = 1 Additional Thread / 3 and + = 5 Additional Threads)
- **uploadsysteme**: Counter of valid contributions (system media) proposed by the user
- **uploadinfos**: Counter of valid contributions (text info) proposed by the user
- **romasso**: Counter of valid contributions (ROM association) proposed by the user
- **uploadmedia**: Counter of valid contributions (game media) proposed by the user
- **maxthreads**: Number of threads allowed for the user (also indicated for non-registered users)
- **maxdownloadspeed**: Download speed (in KB/s) allowed for the user (also indicated for non-registered users)
- **requeststoday**: Total number of API calls during the current day
- **requestskotoday**: Number of API calls with negative return (ROM/game not found) during the current day
- **maxrequestsperdmin**: Maximum number of API calls allowed per minute for the user (see FAQ)
- **maxrequestsperday**: Maximum number of API calls allowed per day for the user (see FAQ)
- **maxrequestskoperday**: Maximum number of API calls with negative return (ROM/game not found) allowed per day for the user (see FAQ)
- **visites**: Number of user visits to ScreenScraper
- **datedernierevisite**: Date of the user's last visit to ScreenScraper (format: yyyy-mm-dd hh:mm:ss)
- **favregion**: Favorite region of user visits to ScreenScraper (france, europe, usa, japan)

**Item XML: jeux** (table) then **jeu**

**Item JSON: jeux** (table)

**Returned Elements**: Identical to jeuInfos API but without ROM information

**Example call:**

```
https://api.screenscraper.fr/api2/jeuRecherche.php?devid=xxx&devpassword=yyy&softname=zzz&output=xml&ssid=test&sspassword=test&systemeid=1&recherche=sonic
```

---

## jeuInfos.php: Game Information / Game Media

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **output**: xml (default), json
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password
- **crc\***: CRC calculation of existing local rom/iso/folder file
- **md5\***: MD5 calculation of existing local rom/iso/folder file
- **sha1\***: SHA1 calculation of existing local rom/iso/folder file
- **systemeid**: Numeric identifier of the system (see systemesListe.php)
- **romtype**: Type of "rom": single rom file / single iso file / folder
- **romnom**: Name of the file (with extension) or folder name
- **romtaille\***: Size in bytes of the file or folder
- **serialnum**: Force game search with the ROM (iso) serial number
- **gameid\*\***: Force game search with its numeric identifier

\* Unless exempted, you must send at least one (preferably all 3) of these calculations (crc, md5, sha1) for rom/iso file or folder identification with your request AND the size (in bytes of the file or folder).

\*\* No ROM information is sent in this case.

### Returned Elements

**Item: serveurs** (ScreenScraper server information)

- **serveurcpu1**: % CPU usage on the current main server
- **serveurcpu2**: % CPU usage on the current secondary server
- **threadsmin**: Number of API accesses in the last 60 seconds
- **nbscrapeurs**: Number of API users currently scraping
- **apiacces**: Number of API accesses today (French time)

**Item: ssuser** (ScreenScraper user information)

- **id**: User's username on ScreenScraper
- **niveau**: User's level on ScreenScraper
- **contribution**: Financial contribution level on ScreenScraper (2 = 1 Additional Thread / 3 and + = 5 Additional Threads)
- **uploadsysteme**: Counter of valid contributions (system media) proposed by the user
- **uploadinfos**: Counter of valid contributions (text info) proposed by the user
- **romasso**: Counter of valid contributions (ROM association) proposed by the user
- **uploadmedia**: Counter of valid contributions (game media) proposed by the user
- **maxthreads**: Number of threads allowed for the user (also indicated for non-registered users)
- **maxdownloadspeed**: Download speed (in KB/s) allowed for the user (also indicated for non-registered users)
- **requeststoday**: Total number of API calls during the current day
- **requestskotoday**: Number of API calls with negative return (ROM/game not found) during the current day
- **maxrequestsperdmin**: Maximum number of API calls allowed per minute for the user (see FAQ)
- **maxrequestsperday**: Maximum number of API calls allowed per day for the user (see FAQ)
- **maxrequestskoperday**: Maximum number of API calls with negative return (ROM/game not found) allowed per day for the user (see FAQ)
- **visites**: Number of user visits to ScreenScraper
- **datedernierevisite**: Date of the user's last visit to ScreenScraper (format: yyyy-mm-dd hh:mm:ss)
- **favregion**: Favorite region of user visits to ScreenScraper (france, europe, usa, japan)

**Item: jeu**

- **id**: Numeric identifier of the game
- **romid**: Numeric identifier of the ROM
- **notgame**: (true/false) indicates if the ROM is assigned to a game or a NON-game (demo/app/...)
- **nom**: Game name (internal ScreenScraper)
- **noms**:
  - **nom_ss**: Game name (internal ScreenScraper)
  - **nom_xx**: Game title region xx (xx = "nomcourt" variable from regionsListe.php API)
- **regionshortnames**:
  - **regionshortname**: Short name of the ROM region (if available)
- **cloneof**: Clone ID (if available)
- **systeme**:
  - **id**: Numeric identifier of the game system
  - **nom**: Name of the game system
  - **parentid**: Numeric identifier of the parent system of the game system
- **editeur**: Publisher name
- **editeurmedias**:
  - **editeurmedia_pictomonochrome**: Media download URL: Publisher Monochrome Logo
  - **editeurmedia_pictocouleur**: Media download URL: Publisher Color Logo
- **developpeur**: Developer name
- **developpeurmedias**:
  - **developpeurmedia_pictomonochrome**: Media download URL: Developer Monochrome Logo
  - **developpeurmedia_pictocouleur**: Media download URL: Developer Color Logo
- **joueurs**: Number of players
- **joueursmedias**:
  - **joueursmedia_pictoliste**: Media download URL: List Pictogram
  - **joueursmedia_pictomonochrome**: Media download URL: Monochrome Pictogram
  - **joueursmedia_pictocouleur**: Media download URL: Color Pictogram
- **note**: Rating out of 20
- **notemedias**:
  - **notemedia_pictoliste**: Media download URL: List Pictogram
  - **notemedia_pictomonochrome**: Media download URL: Monochrome Logo
  - **notemedia_pictocouleur**: Media download URL: Color Logo
- **topstaff**: Game included in ScreenScraper TOP Staff (0: not included, 1: included)
- **rotation**: Game screen rotation (arcade games only)
- **resolution**: Game resolution (arcade games only)
- **synopsis**:
  - **synopsis_xx**: Game description in language xx (xx = "nomcourt" variable from languesListe.php API)
- **classifications**:
  - **classifications\__organisme_**: Game classification
  - **classifications\__organisme_\_medias**: Classification media URLs
- **dates**:
  - **date_xx**: Release date in region xx (france=date_fr, europe=date_eu, usa=date_us, japan=date_jp, ...)
- **genres**: Genre information with IDs, names in multiple languages, and media URLs
- **modes**: Game mode information with IDs, names in multiple languages, and media URLs
- **familles**: Game family information with IDs, names in multiple languages, and media URLs
- **numeros**: Game number in series information with IDs, names in multiple languages, and media URLs
- **themes**: Game theme information with IDs, names in multiple languages, and media URLs
- **styles**: Game style information with IDs, names in multiple languages, and media URLs
- **sp2kcfg**: Text content of .p2k config file (Pad2Keyboard)
- **recalbox actions**: Action information with controls and standardized button text
- **couleurs**: Color information with control identifiers and hexadecimal codes
- **medias**: Extensive media structure including:
  - **media_screenshot**: Screenshot download URL
  - **media_fanart**: Fanart download URL (custom background)
  - **media_video**: Game video capture download URL
  - **media_marquee**: Marquee download URL
  - **media_screenmarquee**: Screen Marquee download URL
  - **media_wheels**: Wheel logos by region
  - **media_wheelscarbon**: Carbon wheel versions by region
  - **media_wheelssteel**: Steel wheel versions by region
  - **media_boitiers**: Box art (texture, 2D, 3D) by region
  - **media_supports**: Support images (texture, 2D) by region and support number
  - **media_flyer**: Flyer images by region and page number
  - **media_manuel**: Manual PDFs by region
  - **media_bezels**: Bezel images (4:3, 16:9, 16:10) by region
- **roms**: List of known ROMs associated with the game, including:
  - **id**: Numeric identifier of the ROM
  - **romnumsupport**: Support number (e.g., 1 = disk 01 or CD 01)
  - **romtotalsupport**: Total number of supports (e.g., 2 = 2 disks or 2 CDs)
  - **romfilename**: ROM file or folder name
  - **romsize**: Size in bytes of the ROM file or folder content
  - **romcrc**: CRC32 calculation result
  - **rommd5**: MD5 calculation result
  - **romsha1**: SHA1 calculation result
  - **romcloneof**: Numeric identifier of parent ROM if clone (Arcade Systems)
  - **beta**: Beta version (0=no / 1=yes)
  - **demo**: Demo version (0=no / 1=yes)
  - **trad**: Translated version (0=no / 1=yes)
  - **hack**: Modified version (0=no / 1=yes)
  - **unl**: Unofficial game (0=no / 1=yes)
  - **alt**: Alternative version (0=no / 1=yes)
  - **best**: Best version (0=no / 1=yes)
  - **netplay**: Netplay compatible (0=no / 1=yes)
- **rom**: Information about the scraped ROM (if found in database) - includes all above fields plus:
  - **romserial**: Manufacturer serial number
  - **romregions**: ROM or folder region(s) (e.g., "fr,us,sp")
  - **romlangues**: ROM or folder language(s) (e.g., "fr,en,es")
  - **romtype**: ROM type
  - **romsupporttype**: Support type
  - **joueurs**: Number of players specific to ROM (if different from "original" game)
  - **dates**: Release dates specific to ROM by region
  - **editeur**: Publisher specific to ROM
  - **developpeur**: Developer specific to ROM
  - **synopsis**: Description specific to ROM by language
  - **clonetypes**: Clone type information
  - **hacktypes**: Hack type information

**Example call:**

```
https://api.screenscraper.fr/api2/jeuInfos.php?devid=xxx&devpassword=yyy&softname=zzz&output=xml&ssid=test&sspassword=test&crc=50ABC90A&systemeid=1&romtype=rom&romnom=Sonic%20The%20Hedgehog%202%20(World).zip&romtaille=749652
```

---

## mediaJeu.php: Download Game Image Media

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password
- **crc**: CRC calculation of existing local image
- **md5**: MD5 calculation of existing local image
- **sha1**: SHA1 calculation of existing local image
- **systemeid**: Numeric identifier of the system (see systemesListe.php)
- **jeuid**: Numeric identifier of the game (see jeuInfos.php)
- **media**: Text identifier of the media to return (see jeuInfos.php)
- **mediaformat** (optional): Media format (extension): ex: jpg, png, mp4, zip, mp3, ... (informative: does not return the media in the specified format)

### Output Parameters

- **maxwidth** (optional): Maximum width in pixels of the returned image
- **maxheight** (optional): Maximum height in pixels of the returned image
- **outputformat** (optional): Format (extension) of the returned image: png or jpg

### Returned Element

PNG Image or Text CRCOK or MD5OK or SHA1OK if the crc, md5 or sha1 parameter is identical to the crc, md5 or sha1 calculation of the server image (update optimization) or Text NOMEDIA if the media file was not found

**Example call:**

```
https://api.screenscraper.fr/api2/mediaJeu.php?devid=xxx&devpassword=yyy&softname=zzz&ssid=test&sspassword=test&crc=&md5=&sha1=&systemeid=1&jeuid=3&media=wheel-hd(wor)
```

---

## mediaVideoJeu.php: Download Game Video Media

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password
- **crc**: CRC calculation of existing local video
- **md5**: MD5 calculation of existing local video
- **sha1**: SHA1 calculation of existing local video
- **systemeid**: Numeric identifier of the system (see systemesListe.php)
- **jeuid**: Numeric identifier of the system (see jeuInfos.php)
- **media**: Text identifier of the media to return (see jeuInfos.php)
- **mediaformat**: Media format (extension): ex: jpg, png, mp4, zip, mp3, ... (optional, informative: does not return the media in the specified format)

### Returned Element

MP4 Video or Text CRCOK or MD5OK or SHA1OK if the crc, md5 or sha1 parameter is identical to the crc, md5 or sha1 calculation of the server video (update optimization) or Text NOMEDIA if the media file was not found

**Example call:**

```
https://api.screenscraper.fr/api2/mediaVideoJeu.php?devid=xxx&devpassword=yyy&softname=zzz&ssid=test&sspassword=test&crc=&md5=&sha1=&systemeid=1&jeuid=3&media=video
```

---

## mediaManuelJeu.php: Download Game Manuals

### Input Parameters

- **devid**: Your developer identifier
- **devpassword**: Your developer password
- **softname**: Name of the calling software
- **ssid** (optional): ScreenScraper user identifier
- **sspassword** (optional): ScreenScraper user password
- **crc**: CRC calculation of existing local manual file
- **md5**: MD5 calculation of existing local manual file
- **sha1**: SHA1 calculation of existing local manual file
- **systemeid**: Numeric identifier of the system (see systemesListe.php)
- **jeuid**: Numeric identifier of the system (see jeuInfos.php)
- **media**: Text identifier of the media to return (see jeuInfos.php)
- **mediaformat**: Media format (extension): ex: jpg, png, mp4, zip, mp3, pdf, ... (optional, informative: does not return the media in the specified format)

### Returned Element

PDF Manual or Text CRCOK or MD5OK or SHA1OK if the crc, md5 or sha1 parameter is identical to the crc, md5 or sha1 calculation of the server manual (update optimization) or Text NOMEDIA if the media file was not found

**Example call:**

```
https://api.screenscraper.fr/api2/mediaManuelJeu.php?devid=xxx&devpassword=yyy&softname=zzz&ssid=test&sspassword=test&crc=&md5=&sha1=&systemeid=1&jeuid=3&media=manuel(eu)
```

---

## botNote.php: System for Automated Sending of Game Ratings

The request must be sent as an HTML request with the "GET" method.

### Parameters

- **ssid**: (type "text") ScreenScraper user identifier
- **sspassword**: (type "text") ScreenScraper user password
- **gameid**: (type "text") Numeric identifier of the game on ScreenScraper
- **note**: (type "integer from 1 to 20") Rating out of 20 for the game

### Returned Element

Returns textual information about the success or failure of the procedure

**Example call:**

```
https://api.screenscraper.fr/api2/botNote.php?devid=xxx&devpassword=yyy&softname=zzz&ssid=test&sspassword=test&gameid=3&note=18
```

---

## botProposition.php: System for Automated Sending of Info or Media Proposals

The request must be sent as an HTML form of type "multipart/form-data" with the "POST" method.

### Input Parameters

- **ssid**: (type "text") ScreenScraper user identifier
- **sspassword**: (type "text") ScreenScraper user password
- **gameid**: (type "text") Numeric identifier of the game on ScreenScraper OR
- **romid**: (type "text") Numeric identifier of the ROM on ScreenScraper

**To propose textual info:**

- **modiftypeinfo**: (type "text") Type of info sent (see "modiftypeinfo" list)
- **modifregion**: (type "text") Short name of the info region (optional: see "modiftypeinfo" list / regions list "regionsListe.php")
- **modiflangue**: (type "text") Short name of the info language (optional: see "modiftypeinfo" list / languages list "languesListe.php")
- **modifversion**: (type "text") Version of the info (optional: see "modiftypeinfo" list)
- **modiftexte**: (type "text") The information itself
- **modifsource**: (type "text") (optional) Source (URL of web page, scan of original support, author, ...) of the information

**To propose media:**

- **modiftypemedia**: (type "text") Type of media sent (see "modiftypemedia" list)
- **modifmediafile**: (type "file") File (File format: see "modiftypemedia" list)
- **modifmediafileurl**: (type "text") URL of media to download (File format: see "modiftypemedia" list)
- **modiftyperegion**: (type "text") Short name of the info region (optional: see "modiftypemedia" list / regions list "regionsListe.php")
- **modiftypenumsupport**: (type "text") Support number (optional: see "modiftypemedia" list / number from 0 to 10)
- **modiftypeversion**: (type "text") Version of the info (optional: see "modiftypemedia" list)
- **modifmediasource**: (type "text") (optional) Source (URL of web page, scan of original support, author, ...) of the information

### Returned Elements

Returns textual information about the success or failure of the procedure

---

## List of Textual Info Types for Games (modiftypeinfo)

| Type | Designation | Region | Language | Multiple Choice | Format |
| --- | --- | --- | --- | --- | --- |
| name | Game name (by Region) | required |  |  | Text |
| editeur | Publisher |  |  |  | Text |
| developpeur | Developer |  |  |  | Text |
| players | Number of Player(s) |  |  |  | Group name (see groups) |
| score | Rating |  |  |  | Rating out of 20 from 0 to 20 |
| rating | Classification |  |  | yes | Group name (see groups) |
| genres | Genre(s) |  |  | yes | French name of group (see groups) |
| datessortie | Release date(s) | required |  |  | Format: yyyy-mm-dd ("xxxx-01-01" if year only) |
| rotation | Rotation |  |  |  | 0 to 360 |
| resolution | Resolution |  |  |  | Text: Width x height in pixels (e.g., 320x240) |
| modes | Game Mode(s) |  |  | yes | Text |
| familles | Family(ies) |  |  | yes | Family name (e.g., "Sonic" for "Sonic 3D pinball") |
| numero | Number |  |  | yes | Series name + Number in series (e.g., "sonic the Hedgehog 2") |
| styles | Style(s) |  |  | yes | Text: Graphic style |
| themes | Theme(s) |  |  |  | Text: Game theme (e.g., "Vampire") |
| description | Synopsis |  | required |  | Text |

---

## List of Textual Info Types for ROMs (modiftypeinfo)

| Type | Designation | Region | Language | Multiple Choice | Format |
| --- | --- | --- | --- | --- | --- |
| developpeur | Developer |  |  |  | Text |
| editeur | Publisher |  |  |  | Text |
| datessortie | Release date(s) | required |  |  | Format: yyyy-mm-dd ("xxxx-01-01" if year only) |
| players | Number of Player(s) |  |  |  | Group name (see groups) |
| regions | Region(s) |  |  | yes | French name of group (see groups) |
| langues | Language(s) |  |  | yes | French name of group (see groups) |
| clonetype | Clone Type(s) |  |  | yes | Text |
| hacktype | Hack Type(s) |  |  | yes | Text |
| friendly | Friendly with... |  |  | yes | Text |
| serial | Manufacturer serial number |  |  | yes | Text |
| description | Synopsis |  | required |  | Text |

---

## List of Media Types (modiftypemedia)

| Type | Designation | Format | Region | Support Num | Multi-Version |
| --- | --- | --- | --- | --- | --- |
| sstitle | Screenshot Title | jpg | required |  |  |
| ss | Screenshot | jpg | required |  |  |
| fanart | Fan Art | jpg |  |  |  |
| video | Video | mp4 |  |  |  |
| overlay | Overlay | png | required |  |  |
| steamgrid | Steam Grid | jpg |  |  |  |
| wheel | Wheel | png | required |  |  |
| wheel-hd | HD Logos | png | required |  |  |
| marquee | Marquee | png |  |  |  |
| screenmarquee | ScreenMarquee | png | required |  |  |
| box-2D | Box: Front | png | required | required |  |
| box-2D-side | Box: Spine | png | required | required |  |
| box-2D-back | Box: Back | png | required | required |  |
| box-texture | Box: Texture | png | required | required |  |
| manuel | Manual | pdf | required |  |  |
| flyer | Flyer | jpg | required | required |  |
| maps | Maps | jpg |  |  | yes |
| figurine | Figurine | png |  |  |  |
| support-texture | Support: Texture | png | required | required |  |
| box-scan | Box: Source(s) | png | required | required | yes |
| support-scan | Support: Source(s) | png | required | required | yes |
| bezel-4-3 | Bezel 4:3 Horizontal | png | required |  |  |
| bezel-4-3-v | Bezel 4:3 Vertical | png | required |  |  |
| bezel-4-3-cocktail | Bezel 4:3 Cocktail | png | required |  |  |
| bezel-16-9 | Bezel 16:9 Horizontal | png | required |  |  |
| bezel-16-9-v | Bezel 16:9 Vertical | png | required |  |  |
| bezel-16-9-cocktail | Bezel 16:9 Cocktail | png | required |  |  |
| wheel-tarcisios | Wheel Tarcisio's | png |  |  |  |
| videotable | Video Table (FullHD) | mp4 |  |  |  |
| videotable4k | Video Table (4k) | mp4 |  |  |  |
| videofronton16-9 | Video Fronton (3 Screens) | mp4 |  |  |  |
| videofronton4-3 | Video Fronton (2 Screens) | mp4 |  |  |  |
| videodmd | Video DMD | mp4 |  |  |  |
| videotopper | Video Topper | mp4 |  |  |  |
| sstable | Screenshot Table | jpg |  |  |  |
| ssfronton1-1 | Screenshot Fronton 1:1 | jpg |  |  |  |
| ssfronton4-3 | Screenshot Fronton 4:3 | jpg |  |  |  |
| ssfronton16-9 | Screenshot Fronton 16:9 | jpg |  |  |  |
| ssdmd | Screenshot DMD | jpg |  |  |  |
| sstopper | Screenshot Topper | jpg |  |  |  |
| themehs | Hyperspin Theme | zip |  |  |  |
| themehb | HyperBat Theme | zip |  |  |  |
