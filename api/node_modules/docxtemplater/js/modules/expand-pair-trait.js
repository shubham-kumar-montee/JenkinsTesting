"use strict";

var traitName = "expandPair";
var mergeSort = require("../mergesort");
var DocUtils = require("../doc-utils");
var wrapper = require("../module-wrapper");

var _require = require("../traits"),
    getExpandToDefault = _require.getExpandToDefault;

var _require2 = require("../errors"),
    getUnmatchedLoopException = _require2.getUnmatchedLoopException,
    getClosingTagNotMatchOpeningTag = _require2.getClosingTagNotMatchOpeningTag,
    throwLocationInvalid = _require2.throwLocationInvalid;

function getOpenCountChange(part) {
	switch (part.location) {
		case "start":
			return 1;
		case "end":
			return -1;
		default:
			throwLocationInvalid(part);
	}
}

function getPairs(traits) {
	var errors = [];
	var pairs = [];
	if (traits.length === 0) {
		return { pairs: pairs, errors: errors };
	}
	var countOpen = 1;
	var firstTrait = traits[0];
	if (firstTrait.part.location === "start") {
		for (var i = 1; i < traits.length; i++) {
			var currentTrait = traits[i];
			countOpen += getOpenCountChange(currentTrait.part);
			if (countOpen === 0) {
				var _outer = getPairs(traits.slice(i + 1));
				if (currentTrait.part.value !== firstTrait.part.value && currentTrait.part.value !== "") {
					errors.push(getClosingTagNotMatchOpeningTag({ tags: [firstTrait.part, currentTrait.part] }));
				} else {
					pairs = [[firstTrait, currentTrait]];
				}
				return { pairs: pairs.concat(_outer.pairs), errors: errors.concat(_outer.errors) };
			}
		}
	}
	var part = firstTrait.part;
	errors.push(getUnmatchedLoopException({ part: part, location: part.location }));
	var outer = getPairs(traits.slice(1));
	return { pairs: outer.pairs, errors: errors.concat(outer.errors) };
}

var expandPairTrait = {
	name: "ExpandPairTrait",
	postparse: function postparse(postparsed, _ref) {
		var getTraits = _ref.getTraits,
		    _postparse = _ref.postparse;

		var traits = getTraits(traitName, postparsed);
		traits = traits.map(function (trait) {
			return trait || [];
		});
		traits = mergeSort(traits);

		var _getPairs = getPairs(traits),
		    pairs = _getPairs.pairs,
		    errors = _getPairs.errors;

		var expandedPairs = pairs.map(function (pair) {
			var expandTo = pair[0].part.expandTo;
			if (expandTo === "auto") {
				expandTo = getExpandToDefault(postparsed.slice(pair[0].offset, pair[1].offset));
			}
			if (!expandTo) {
				return [pair[0].offset, pair[1].offset];
			}
			var left = DocUtils.getLeft(postparsed, expandTo, pair[0].offset);
			var right = DocUtils.getRight(postparsed, expandTo, pair[1].offset);
			return [left, right];
		});

		var currentPairIndex = 0;
		var innerParts = void 0;
		var newParsed = postparsed.reduce(function (newParsed, part, i) {
			var inPair = currentPairIndex < pairs.length && expandedPairs[currentPairIndex][0] <= i;
			var pair = pairs[currentPairIndex];
			var expandedPair = expandedPairs[currentPairIndex];
			if (!inPair) {
				newParsed.push(part);
				return newParsed;
			}
			if (expandedPair[0] === i) {
				innerParts = [];
			}
			if (pair[0].offset !== i && pair[1].offset !== i) {
				innerParts.push(part);
			}
			if (expandedPair[1] === i) {
				var basePart = postparsed[pair[0].offset];
				delete basePart.location;
				delete basePart.expandTo;
				basePart.subparsed = _postparse(innerParts);
				newParsed.push(basePart);
				currentPairIndex++;
			}
			return newParsed;
		}, []);
		return { postparsed: newParsed, errors: errors };
	}
};

module.exports = function () {
	return wrapper(expandPairTrait);
};