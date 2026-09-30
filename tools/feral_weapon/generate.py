#!/usr/bin/env python3
"""Build candidate physical-weapon inputs. Caster adjustments need sheet benchmarks.

Usage: python3 tools/feral_weapon/generate.py FOREVER_ITEMSPARSE ERA_ITEMSPARSE
The two CSV inputs are from wago.tools, builds 1.60.1.70124 and 1.15.9.69722.
"""
import csv, hashlib, json, pathlib, sys
root = pathlib.Path(__file__).resolve().parents[2]
source = root / 'assets/database/db.json'
db = json.loads(source.read_text())
rows = {}
for filename in [sys.argv[2], sys.argv[1]]:
    rows.update({int(row['ID']): row for row in csv.DictReader(open(filename))})
tables = {}
for hand in ['OneHand', 'TwoHand']:
    tables[hand] = {int(r['ItemLevel']): r for r in csv.DictReader(open(root / f'tools/feral_weapon/ItemDamage{hand}-70124.csv'))}
weapons = {}
for item in db['items']:
    if item.get('type') != 13 or item.get('handType') not in [1, 2, 4] or not item.get('weaponSpeed'):
        continue
    ident = item['id']; props = item['scalingOptions']['0']; level = props['ilvl']; quality = item.get('quality', 0)
    record = {'name': item['name'], 'itemLevel': level, 'quality': quality, 'hand': 'TwoHand' if item['handType'] == 4 else 'OneHand'}
    row = rows.get(ident)
    # Spell/healing power in old catalog entries is not a reliable caster flag.
    # Exclude those inputs until the supplied sheet identifies their table rule.
    spell_stats = {str(i) for i in range(4, 16)}
    caster = row and (float(row['QualityModifier']) < 0 or int(row['Flags_1']) & 0x200)
    if row is None:
        record['unresolved'] = 'No client ItemSparse record; caster classification unverified'
    elif caster or any(float(v) for k, v in props.get('stats', {}).items() if k in spell_stats):
        record['unresolved'] = 'Caster or spell-stat weapon; requires the community benchmark and effective-DPS rule'
    elif quality > 6 or level not in tables[record['hand']]:
        record['unresolved'] = 'No matching DPS table row'
    else:
        record['tableDps'] = float(tables[record['hand']][level][f'Quality_{quality}'])
    weapons[str(ident)] = record
out = {'clientBuild': '1.60.1.70124', 'catalogSHA256': hashlib.sha256(source.read_bytes()).hexdigest(), 'weapons': weapons}
(root / 'sim/druid/data/feral_weapon_candidates.json').write_text(json.dumps(out, indent=2, sort_keys=True) + '\n')
print('Resolved physical weapons:',sum('tableDps' in w for w in weapons.values()),'unresolved:',sum('unresolved' in w for w in weapons.values()))
