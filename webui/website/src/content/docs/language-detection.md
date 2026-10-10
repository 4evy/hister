---
date: '2026-10-10T00:00:00+02:00'
draft: false
title: 'Language Settings and Detection'
description: 'Choose language detection settings, balance accuracy and memory use, preserve stop words, and apply changes to existing documents.'
---

Hister detects the language of document text so it can apply the appropriate search analyzer. An analyzer splits text into searchable terms and can normalize spelling, reduce related word forms to a common stem, and remove common words. Choosing the right analyzer helps searches match the language of your content.

The defaults enable detection for all supported languages with reduced memory use. Start with these settings unless short text, memory constraints, or phrase searches give you a reason to change them.

## Default Configuration

All four settings belong in the `indexer` section of your [configuration file](/docs/configuration):

```yaml
indexer:
  detect_languages: true
  languages: []
  language_detection_accuracy: 'low'
  keep_stopwords: false
```

| Setting                       | Default | Options and effect                                                                                                  |
| ----------------------------- | ------- | ------------------------------------------------------------------------------------------------------------------- |
| `detect_languages`            | `true`  | `true` enables detection; `false` uses the default analyzer for normally processed documents.                       |
| `languages`                   | `[]`    | An empty list considers all supported languages. A nonempty list needs at least two distinct supported codes.       |
| `language_detection_accuracy` | `low`   | `low` uses less memory; `high` uses additional models for short text. An omitted or empty value also selects `low`. |
| `keep_stopwords`              | `false` | `false` lets language analyzers remove stop words; `true` preserves them while retaining other language analysis.   |

These settings apply on the server and in service imports that process documents in the command line client. Configure each machine that performs that processing. They control document indexing, not the interface language or the language of pages fetched from a website.

## How Detection Affects Search

For documents without a manual language pin, Hister runs Lingua locally on the extracted plain text in the document's `text` field. Detection itself does not call an external service. The result comes from that text, rather than the browser's language preference, the page's HTML `lang` attribute, or its title alone.

Hister assigns one language to the whole document. It does not split a page into separate language regions, so mixed language pages, code, names, and very short snippets can produce an unexpected result.

Languages with an available analyzer use a separate search index, such as `index_en.db` for English. A regional tag such as `en-US` stays in the document metadata but shares the English analyzer and index with `en` and `en-GB`. Norwegian Bokmal (`nb`) uses the Norwegian analyzer in `index_no.db`. Chinese (`zh`), Japanese (`ja`), and Korean (`ko`) use Bleve's shared CJK analyzer in their respective indexes. Searches run across the active indexes, applying their respective analyzers to text queries. You do not need to choose a language before searching.

Documents classified as `unknown`, or without an available language analyzer, use `index.db`. Its default analyzer splits text using Unicode word boundaries and lowercases terms. It keeps stop words and does not apply language specific stemming. The content remains searchable, but different forms of a word may no longer match each other.

Document tags, detection candidates, and analyzers are separate. For example, Indonesian (`id`) can be detected but uses the default analyzer. Ukrainian (`uk`) can be assigned manually and searched by its language tag, even though it has neither a bundled detector model nor a language analyzer.

## Choosing Settings

| Collection or priority              | Suggested settings                                     | Benefits                                                       | Tradeoffs                                                                                                                                   |
| ----------------------------------- | ------------------------------------------------------ | -------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| General browsing across languages   | `languages: []`, accuracy `low`                        | Covers all supported languages while limiting model memory.    | Short snippets can be harder to identify correctly.                                                                                         |
| Content in a known set of languages | A list such as `['en', 'de']`, accuracy `low`          | Loads fewer models and narrows the possible detection results. | Text outside the list can be assigned the wrong language or become `unknown`.                                                               |
| Short notes or snippets             | Accuracy `high`, with all languages or a suitable list | Uses more evidence for short text.                             | Loads more models and uses more memory; correct detection is still not guaranteed.                                                          |
| Minimal detection overhead          | `detect_languages: false`                              | Avoids language model loading for newly processed documents.   | New documents lose automatic language labels and language specific analysis. Existing indexes stay open until reindexing consolidates them. |
| Phrases containing common words     | `keep_stopwords: true` with detection enabled          | Keeps words such as `for` and `your` searchable.               | Retains more indexed terms and can increase index size.                                                                                     |

The accuracy and candidate settings are independent. You can use `high` with a restricted list or `low` with all supported languages. `keep_stopwords` changes analysis after detection; it does not improve the detector's accuracy.

## Accuracy and Memory

`low` uses trigram models, which compare sequences of three letters. `high` can also consult models for sequences of one, two, four, and five letters when the text is short. Both modes use only trigrams once the detector counts at least 120 letters, so the extra models primarily help with short extracted text.

Short text can become `unknown` in low accuracy mode, but this is not a rule that every short document is rejected. Lingua also uses writing system and character rules, which can identify some very short text in either mode. High accuracy can improve short text detection without guaranteeing a correct result.

To favor short text accuracy:

```yaml
indexer:
  detect_languages: true
  languages: []
  language_detection_accuracy: 'high'
```

The low accuracy default also applies to older configuration files that omit this option. Set `high` explicitly if you want the earlier behavior for short text.

Models load as needed and remain cached for the lifetime of the process. A few short documents in high accuracy mode can therefore leave additional models in memory even if most later documents are long. Restricting candidates prevents excluded models from loading; low accuracy avoids the additional model sizes. Restart Hister after changing settings to release models that are already cached.

Memory use depends on the documents processed and the languages considered. Historical [measurements in PR #584](https://github.com/asciimoo/hister/pull/584) illustrate the savings from restricting languages and using low accuracy, but describe one workload. They are neither a memory guarantee nor an accuracy benchmark, and detector memory is only part of Hister's total memory use.

## Restricting Languages

Use `languages` when you know which languages your collection contains:

```yaml
indexer:
  detect_languages: true
  languages: ['en', 'de', 'nl', 'id']
  language_detection_accuracy: 'low'
```

Codes ignore case and surrounding whitespace. Regional tags select the base language's model, so `en-US` selects English without making detection distinguish American English. Duplicates and aliases do not count as distinct candidates: `['en', 'en-GB']` and `['no', 'nb']` are invalid. Languages without bundled detection models are rejected. This validation also applies when detection is disabled. There is no single language mode that forces every document to English, for example.

The list controls detection candidates. It does not reject documents written in other languages or restrict the languages you can search for in existing content.

### Confidence and Unknown Results

For a nonempty list, Hister requires the highest confidence score to be at least `0.8` and greater than the next score. Otherwise it stores `unknown` and uses the default analyzer. The cutoff is fixed, with no configuration option to change it.

With an empty list, Hister uses the highest scoring supported language without the `0.8` cutoff, but still rejects tied or unavailable results. Even explicitly listing every supported language activates the cutoff, so it is not equivalent to `languages: []`.

Confidence scores are relative to the candidates. A score of `0.8` is not an 80 percent guarantee that the classification is correct. Text in an excluded language can receive a high score for an included language, particularly when they share a writing system. Restricting the list cannot guarantee that excluded languages become `unknown`.

Keep the full set if your browsing spans unpredictable languages. An incorrect language analyzer can change which word forms match, and the default analyzer can miss matches that the correct language analyzer would find.

### Supported Codes

Hister bundles detection models for these 30 base language codes. Aliases select the same model as their canonical code and do not count as additional detection candidates.

| Code | Language         | Aliases |
| ---- | ---------------- | ------- |
| `ar` | Arabic           |         |
| `bg` | Bulgarian        |         |
| `ca` | Catalan          |         |
| `da` | Danish           |         |
| `de` | German           |         |
| `el` | Greek            |         |
| `en` | English          |         |
| `es` | Spanish          |         |
| `eu` | Basque           |         |
| `fa` | Persian          |         |
| `fi` | Finnish          |         |
| `fr` | French           |         |
| `ga` | Irish            |         |
| `hi` | Hindi            |         |
| `hr` | Croatian         |         |
| `hu` | Hungarian        |         |
| `hy` | Armenian         |         |
| `id` | Indonesian       | `in`    |
| `it` | Italian          |         |
| `ja` | Japanese         |         |
| `ko` | Korean           |         |
| `nb` | Norwegian Bokmal | `no`    |
| `nl` | Dutch            |         |
| `pl` | Polish           |         |
| `pt` | Portuguese       |         |
| `ro` | Romanian         | `mo`    |
| `ru` | Russian          |         |
| `sv` | Swedish          |         |
| `tr` | Turkish          |         |
| `zh` | Chinese          |         |

Prefer the canonical codes for new configuration. The alias `no` preserves Hister's historical spelling for Norwegian Bokmal; `in` and `mo` are deprecated language codes. Regional and script tags such as `en-US` and `zh-Hant` select their base models. Automatic detection returns a base code, such as `en`, `zh`, or `nb`. The fallback value `unknown` is not a detection candidate.

Manual pins accept known BCP 47 tags, including scripts, regions, and variants, even when no detector model or analyzer is available. Tags are normalized, so `EN-us` is stored as `en-US`. Use hyphens between subtags. Private language codes and tags with extensions are rejected. `unknown` and `und` both select the stored label `unknown`.

Manual pins also accept `iw` for Hebrew (`he`) and `ji` for Yiddish (`yi`). These languages have no bundled detection models, so they are not included in the table.

## Stop Words and Phrase Searches

Stop words are common words that an analyzer may remove because they occur in many documents. By default, language analyzers apply their own stop word filters. For example, the English analyzer turns `for your information` into the indexed term `inform`.

To keep those common words:

```yaml
indexer:
  detect_languages: true
  keep_stopwords: true
```

The English example then produces `for`, `your`, and `inform`. Stemming and other normalization remain active. This is useful for searches such as `text:"for your information"`, where the common words are part of the phrase you want to match. Quoting a phrase still uses analyzed terms, so it does not turn search into an exact comparison of the original characters.

Preserving stop words can increase the number of indexed terms and index size. It has no effect on the default analyzer, which already keeps them, or on a language analyzer that has no stop word filter. It also does not restore words omitted from existing indexes until you reindex.

## Environment Overrides

Environment variables override the corresponding YAML values. A language list uses commas:

```bash
export HISTER__INDEXER__DETECT_LANGUAGES=true
export HISTER__INDEXER__LANGUAGES=en,de,nl,id
export HISTER__INDEXER__LANGUAGE_DETECTION_ACCURACY=low
export HISTER__INDEXER__KEEP_STOPWORDS=false
```

To return to the full candidate set, set `languages: []` in YAML and remove any `HISTER__INDEXER__LANGUAGES` override. See [Environment Variables](/docs/configuration#environment-variables) for configuration precedence.

## Applying Changes to Existing Documents

Restart the server after editing its configuration. A changed candidate list or accuracy mode applies when documents are next processed; it does not relabel the stored collection automatically.

| Changed setting               | Action for existing documents                                                                                                                                                      |
| ----------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `languages`                   | Reindex if you want existing documents classified using the new candidates. Previously created language indexes remain searchable even if their language is removed from the list. |
| `language_detection_accuracy` | Reindex if you want existing documents classified using the new accuracy mode.                                                                                                     |
| `detect_languages`            | Reindex if you want to apply the new detection and analyzer behavior to existing documents. They remain searchable without reindexing.                                             |
| `keep_stopwords`              | Run a full reindex to rebuild the stored analyzer mappings and terms.                                                                                                              |

After restarting with the desired settings, leave the server running and request a rebuild:

```bash
hister reindex
```

Use the same `detect_languages` setting for this command and the server. The command sends its configured value to the server, while the candidate list, accuracy mode, and stop word setting come from the running server. If you use a custom configuration path, pass it with `hister --config /path/to/config.yml reindex`.

When detection is disabled, existing language indexes remain open and searchable with their stored analyzers and language labels. Newly processed documents without a manual language pin use the default analyzer and receive the `unknown` language label. Reprocessing an unpinned document moves it into the default index without leaving a duplicate in its old language index. Run a full reindex if you want to consolidate the unpinned collection into the default index. Manually pinned documents retain their language and analyzer, including when detection is disabled.

Hister records the analyzer configuration and warns at startup when `detect_languages` or the effective `keep_stopwords` setting differs from the stored index configuration. Changes to the candidate list or accuracy mode do not trigger that warning. Reindexing also performs normal document processing and rule checks; see [Data Storage and Lifecycle](/docs/data-lifecycle) for its broader effects.

## Checking and Correcting Languages

Use the [query language](/docs/query-language) to inspect stored classifications:

```textplain
language:en
language:en-US
language:(en|de)
language:unknown
```

Filters ignore case and include more specific tags: `language:en` matches `en`, `en-US`, and `en-GB`, while `language:en-US` selects the American English tag and any more specific variants. Both `language:nb` and `language:no` find old `no` labels and new `nb` labels. Existing index files remain searchable without rebuilding. Reprocessing normalizes old labels, so language facets can show both spellings until those documents are updated or reindexed.

`language:unknown` includes documents for which detection was disabled or inconclusive, and documents manually pinned to `unknown`. It is not a reliable list of all documents in unsupported or excluded languages, since those can be misclassified.

You can pin a stored language through the update command, even when detection is disabled. For example, to mark the documents on a known English site as English:

```bash
hister update 'domain:english.example.com' --language en --dry
hister update 'domain:english.example.com' --language en
```

Use a regional tag such as `--language en-US` when that distinction is known. Detection does not infer regional tags.

The first command reports the affected count; the second asks for confirmation and updates the matching documents. A language change rebuilds those documents in the appropriate index without requiring a full reindex. The choice remains pinned across server restarts, reindexing, browser resubmissions, imports, and watched file updates. Pins belong to a document and its owner, so correcting your copy does not affect another user's copy of the same URL. Pinning a document to its current language still counts as an update if it was not already pinned.

Use `--language unknown` to pin the default analyzer. To clear a pin and classify the stored text using the running server's detection settings:

```bash
hister update 'domain:english.example.com' --language auto --dry
hister update 'domain:english.example.com' --language auto
```

With detection disabled, `auto` clears the pin and sets the language to `unknown`. Future processing follows the configured detector again. This operation uses stored text without fetching the page or rereading the source file. A dry run leaves the pin and document unchanged.

The JSON document field `language_override` records the pin separately from the resulting `language`. Exports and imports preserve it. An ordinary resubmission cannot replace or clear an existing pin; use `hister update` or `/api/update` for that. The API accepts `{"language":"auto"}` in the update request's `changes` object.

Language corrections made before persistent pins were introduced have no saved override marker. Reapply those corrections once to pin them; existing language labels alone are not treated as manual choices.
