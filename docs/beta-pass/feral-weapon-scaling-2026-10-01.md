# Feral weapon scaling candidate, October 1, 2026

Status: candidate only, not pinned or deployed. The community formula supplied by Sage now has an implementation and regression coverage. Production retains f3ebe9fffbe721b3fe739179ec3389a6706008be until the linked spreadsheet's effective-DPS rules and in-game benchmark rows are available.

## Formula implementation

`sim/core/feral_weapon.go` converts each input, constant and intermediate result to float32. Multiplication retains the supplied order, with swing times of 1000 ms for Cat and 2500 ms for Bear. Armed variance is 0.4. The lower pre-AP endpoint rounds down; the upper rounds to nearest; both clamp to at least 1. Unarmed uses 0.5 DPS and zero variance. The AP contribution retains `(AP * (0.001 / 14)) * milliseconds` in float32.

The source describes the final floor and ceiling as character-sheet display rounding. `FeralWeaponCharacterSheetDamage` implements that separately. Combat retains the fractional AP contribution rather than treating the display endpoints as quantized combat rolls. The existing random damage model is otherwise preserved. Confirm this distinction with the spreadsheet and combat evidence before adoption.

Both form weapons read the equipped weapon's exact table DPS. Normalized specials keep form speed, not the equipped weapon's speed. Existing Feral Attack Power and normal AP continue through the existing stat dependencies. Unrelated classes keep their original weapon calculations. Re-entering a form recomputes the weapon instead of restoring a value captured during initialization.

## Client inputs

Sources:

- https://wago.tools/db2/ItemDamageOneHand/csv?build=1.60.1.70124
- https://wago.tools/db2/ItemDamageTwoHand/csv?build=1.60.1.70124
- https://wago.tools/db2/ItemSparse/csv?build=1.60.1.70124
- https://wago.tools/db2/ItemSparse/csv?build=1.15.9.69722

The two DPS tables are byte-identical between builds 70009 and 70124. Their checksums are saved in `data/feral-weapon-table-provenance-2026-10-01.json`. The Forever ItemSparse table omits some unchanged older items, including Impervious Giant; the Era table supplies their metadata. The engine's final item catalog supplies equipped item level, quality and handedness. No rounded tooltip DPS is used to reconstruct the table value.

The candidate resolves 1,195 physical-weapon inputs from 1,388 melee weapons. It rejects 193 unresolved inputs explicitly. These include 100 without a client ItemSparse record and weapons whose caster classification or effective DPS needs validation. No item-quality downgrade or epic two-thirds multiplier is guessed into production. Spell stats are used only to flag an input for review, not to assert that it has a caster flag.

As a separate input sanity check, the formula reproduces 1,182 of those 1,195 physical weapon tooltip ranges with the stored weapon speed and client variance. The remaining 13 have catalog/client discrepancies or special weapon damage that need individual review. This is not a substitute for in-game form-damage benchmarks.

Rebuild the candidate data with:

```sh
python3 tools/feral_weapon/generate.py /path/to/ItemSparse-70124.csv /path/to/ItemSparse-era.csv
```

## Verification

Passed:

```sh
go test -tags with_db ./sim/core -count=1
go test -tags with_db ./sim/core ./sim/druid/feralcat ./sim/druid/feralbear \
  -run 'TestFeralWeapon|TestWolfshead|Test.*Windfury' -count=1
go build -tags with_db -o /tmp/feral-weapon-scaling/wowsimcli ./cmd/wowsimcli
```

Formula cases cover both forms, naked weapons, the sub-one-DPS clamp, fractional AP and float32 rounding boundaries where a float64 shortcut changes the result. Integration covers equipped weapon changes and Cat/Bear re-entry. Ordinary weapon arithmetic remains unchanged.

The current production baseline requests were replayed unchanged at 10,000 iterations and their saved seeds:

| Reference | Weapon | Live baseline | Candidate |
| --- | --- | ---: | ---: |
| Cat | Heartseeker, table DPS 41.49628067017 | 511.325048 | 486.384776 |
| Bear | Impervious Giant, table DPS 49.02209854126 | 404.955763 | 392.309643 |

These are candidate implementation results, not verified beta predictions or updated rankings. The old fixed paws effectively granted more base DPS than either equipped weapon. Weapon selection must be re-optimized after validation because weapon DPS now matters.

## Before release

1. Obtain the spreadsheet referenced in the community post, especially its in-game/API benchmark rows and caster cases.
2. Validate the formula against those rows, including naked forms and fishing poles. Confirm combat versus final character-sheet rounding.
3. Resolve effective DPS and classification for caster weapons. The post's uncommon/rare quality downgrade and high-level epic two-thirds observation do not specify a universal rule or an epic level boundary.
4. Replace this intentionally incomplete, rejecting candidate catalog with the validated policy; wire regeneration into catalog updates and add drift checks.
5. Re-optimize the Cat and Bear reference weapons, update expected integration results, and complete the coordinated engine/catalog/worker release. Refresh affected rankings and both gear-pool comparisons with saved live evidence.
