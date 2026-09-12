# Concept: Lenient Parsing and Last-Write-Wins

**Audience:** everyone editing GSL or building with it. **Prerequisite:** [Parents, scopes and edge dependencies](parents-and-scopes.md).
**Next step:** [Canonical form and diffability](canonical-form.md).

GSL's parser is deliberately **lenient** and its declarations follow **last-write-wins**. These two rules are why text-only graphs survive concurrent editing without ceremony — and why reading a warning is a normal part of using the language.

## Lenient: warnings, not walls

Unknown syntax, duplicate member declarations, out-of-order sets — the parser reports these as **warnings, not errors**. The graph still builds; nothing you already wrote is thrown away for a warning you might not care about. In Go, `Parse` returns `(graph, warnings, err)`: only genuine corruption produces `err`, and you are expected to *check the warnings too* ([Go reference](../../GO_REFERENCE.md#parse-correctly)).

What GSL does **not** do is police your graph's *meaning*:

- No schema validation — a node inside a parent or a multi-homed child is your model's concern, not the parser's.
- No acyclicity, tree-validity or structural checks — graphs are accepted as written.
- No attribute types — anything in brackets is stored, and you interpret it (see [Untyped attributes](attributes.md)).

That leniency is a deployment feature: a new attribute or an unusual-but-harmless shape never blocks anyone.

## Last-write-wins: merging without merge rituals

When the same thing is declared more than once — the same node twice, the same node in two parent blocks — GSL does not error. It keeps the **last** declaration's attributes and membership:

```gsl
node orders [team="orders"]
node orders [team="orders", replicas="6"]
```

The second line simply wins. This turns concurrent edits into a tractable rule: **later declarations replace earlier ones**, and the final file is the file. Git merges of GSL tend to be small, textual and unsurprising — exactly the property a canonical format exists to provide.

## Two rules, one behaviour

Punishment-free parse + deterministic overwrite = a format you can resolve by hand in a merge, not by regenerating a binary artifact. The trade-off, honestly stated: leniency means a typo can silently fall through as a warning, so *check your warnings* when it matters — a release gate that validates `gsl-query` round-trips are clean is a natural place to enforce it (see [Render a view for CI](../cookbook/render-a-view-for-ci.md)).

---

**Next:** [Canonical form and diffability](canonical-form.md) — the property that makes "text graph" equal "sortable, reviewable asset".