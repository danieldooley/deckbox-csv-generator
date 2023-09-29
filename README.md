


# Card Codes
When in card adding mode there is a special syntax for adding codes. It roughly looks like:
```
([SET CODE].)[CARD SET NUMBER](f)
```

This means:
* `([SET CODE].)` - An optional set code (like `CMM` for commander masters) followed by a literal `.`
* `[CARD SET NUMBER]` - The number of the card within the set, leading zeros will be removed.
* `(f)` - An optional literal `f` indicating the card is a foil.

If the set code is excluded the top level set selected will be used. If manually entering the set code you don't need
to set one at the top level.

## Examples:

```
CMM.0694f
```
Will enter a foil Fierce Guardianship (card number `694`) from Commander Masters.

```
357
```
With `Throne of Eldraine` selected as the top level set will add a non-foil Wishclaw Talisman.
