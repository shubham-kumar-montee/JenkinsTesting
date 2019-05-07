"use strict";

var FileTypeConfig = require("../file-type-config.js");
var XmlTemplater = require("../xml-templater");
var path = require("path");
var Docxtemplater = require("../docxtemplater.js");
var DocUtils = Docxtemplater.DocUtils;
var chai = require("chai");
var expect = chai.expect;
var JSZip = require("jszip");
var xmlPrettify = require("./xml-prettify");
var fs = require("fs");

var _require = require("lodash"),
    get = _require.get,
    unset = _require.unset,
    omit = _require.omit,
    uniq = _require.uniq;

var countFiles = 1;
var allStarted = false;
var examplesDirectory = void 0;
var docX = {};
var imageData = {};

/* eslint-disable no-console */

function createXmlTemplaterDocx(content, options) {
	options = options || {};
	options.fileTypeConfig = FileTypeConfig.docx;
	Object.keys(DocUtils.defaults).forEach(function (key) {
		var defaultValue = DocUtils.defaults[key];
		options[key] = options[key] != null ? options[key] : defaultValue;
	});
	options.modules = options.fileTypeConfig.baseModules.map(function (moduleFunction) {
		var module = moduleFunction();
		module.optionsTransformer({}, options);
		return module;
	});

	return new XmlTemplater(content, options).setTags(options.tags).parse();
}

function writeFile(expectedName, zip) {
	var writeFile = path.resolve(examplesDirectory, "..", expectedName);
	if (fs.writeFileSync) {
		fs.writeFileSync(writeFile, zip.generate({ type: "nodebuffer", compression: "DEFLATE" }));
	}
}
function unlinkFile(expectedName) {
	var writeFile = path.resolve(examplesDirectory, "..", expectedName);
	if (fs.unlinkSync) {
		try {
			fs.unlinkSync(writeFile);
		} catch (e) {
			if (e.code !== "ENOENT") {
				throw e;
			}
		}
	}
}

function shouldBeSame(options) {
	var zip = options.doc.getZip();
	var expectedName = options.expectedName;
	var expectedZip = void 0;

	try {
		expectedZip = docX[expectedName].zip;
	} catch (e) {
		writeFile(expectedName, zip);
		console.log(JSON.stringify({ msg: "Expected file does not exists", expectedName: expectedName }));
		throw e;
	}

	try {
		uniq(Object.keys(zip.files).concat(Object.keys(expectedZip.files))).map(function (filePath) {
			var suffix = "for \"" + filePath + "\"";
			expect(expectedZip.files[filePath]).to.be.an("object", "The file " + filePath + " doesn't exist on " + expectedName);
			expect(zip.files[filePath]).to.be.an("object", "The file " + filePath + " doesn't exist on generated file");
			expect(zip.files[filePath].name).to.be.equal(expectedZip.files[filePath].name, "Name differs " + suffix);
			expect(zip.files[filePath].options.dir).to.be.equal(expectedZip.files[filePath].options.dir, "IsDir differs " + suffix);
			var text1 = zip.files[filePath].asText().replace(/\n|\t/g, "");
			var text2 = expectedZip.files[filePath].asText().replace(/\n|\t/g, "");
			if (filePath.indexOf(".png") !== -1) {
				expect(text1.length).to.be.equal(text2.length, "Content differs " + suffix);
				expect(text1).to.be.equal(text2, "Content differs " + suffix);
			} else {
				if (text1 === text2) {
					return;
				}
				var pText1 = xmlPrettify(text1, options);
				var pText2 = xmlPrettify(text2, options);
				expect(pText1).to.be.equal(pText2, "Content differs " + suffix + " lengths: \"" + text1.length + "\", \"" + text2.length + "\"");
			}
		});
	} catch (e) {
		writeFile(expectedName, zip);
		console.log(JSON.stringify({ msg: "Expected file differs from actual file", expectedName: expectedName }));
		throw e;
	}
	unlinkFile(expectedName);
}

function checkLength(e, expectedError, propertyPath) {
	var propertyPathLength = propertyPath + "Length";
	var property = get(e, propertyPath);
	var expectedPropertyLength = get(expectedError, propertyPathLength);
	if (property && expectedPropertyLength) {
		expect(expectedPropertyLength).to.be.a("number", JSON.stringify(expectedError.properties));
		expect(expectedPropertyLength).to.equal(property.length);
		unset(e, propertyPath);
		unset(expectedError, propertyPathLength);
	}
}

function cleanError(e, expectedError) {
	delete e.properties.explanation;
	if (expectedError.properties.offset != null) {
		expect(e.properties.offset).to.be.deep.equal(expectedError.properties.offset);
	}
	delete e.properties.offset;
	delete expectedError.properties.offset;
	e = omit(e, ["line", "sourceURL", "stack"]);
	if (e.properties.postparsed) {
		e.properties.postparsed.forEach(function (p) {
			delete p.lIndex;
			delete p.offset;
		});
	}
	if (e.properties.rootError) {
		expect(e.properties.rootError, JSON.stringify(e.properties)).to.be.instanceOf(Error);
		expect(expectedError.properties.rootError, JSON.stringify(expectedError.properties)).to.be.instanceOf(Object);
		if (expectedError) {
			expect(e.properties.rootError.message).to.equal(expectedError.properties.rootError.message);
		}
		delete e.properties.rootError;
		delete expectedError.properties.rootError;
	}
	checkLength(e, expectedError, "properties.paragraphParts");
	checkLength(e, expectedError, "properties.postparsed");
	if (e.stack && expectedError) {
		expect(e.stack).to.contain("Error: " + expectedError.message);
	}
	delete e.stack;
	return e;
}

function wrapMultiError(error) {
	var type = Object.prototype.toString.call(error);
	var errors = void 0;
	if (type === "[object Array]") {
		errors = error;
	} else {
		errors = [error];
	}

	return {
		name: "TemplateError",
		message: "Multi error",
		properties: {
			id: "multi_error",
			errors: errors
		}
	};
}

function expectToThrow(fn, type, expectedError) {
	var e = null;
	try {
		fn();
	} catch (error) {
		e = error;
	}
	expect(e, "No error has been thrown").not.to.be.equal(null);
	var toShowOnFail = e.stack;
	expect(e, toShowOnFail).to.be.instanceOf(Error);
	expect(e, toShowOnFail).to.be.instanceOf(type);
	expect(e, toShowOnFail).to.be.an("object");
	expect(e, toShowOnFail).to.have.property("properties");
	expect(e.properties, toShowOnFail).to.be.an("object");
	expect(e.properties, toShowOnFail).to.have.property("explanation");
	expect(e.properties.explanation, toShowOnFail).to.be.a("string");
	expect(e.properties, toShowOnFail).to.have.property("id");
	expect(e.properties.id, toShowOnFail).to.be.a("string");
	expect(e.properties.explanation, toShowOnFail).to.be.a("string");
	e = cleanError(e, expectedError);
	if (e.properties.errors) {
		var msg = "expected : \n" + JSON.stringify(expectedError.properties.errors) + "\nactual : \n" + JSON.stringify(e.properties.errors);
		expect(expectedError.properties.errors).to.be.an("array", msg);
		expect(e.properties.errors.length).to.equal(expectedError.properties.errors.length, msg);
		e.properties.errors = e.properties.errors.map(function (e, i) {
			return cleanError(e, expectedError.properties.errors[i]);
		});
	}
	expect(JSON.parse(JSON.stringify(e))).to.be.deep.equal(expectedError);
}

function load(name, content, fileType, obj) {
	var zip = new JSZip(content);
	obj[name] = new Docxtemplater();
	obj[name].loadZip(zip);
	obj[name].loadedName = name;
	obj[name].loadedContent = content;
	return obj[name];
}
function loadDocument(name, content) {
	return load(name, content, "docx", docX);
}
function loadImage(name, content) {
	imageData[name] = content;
}

function loadFile(name, callback) {
	countFiles += 1;
	if (fs.readFileSync) {
		var _path = require("path");
		var buffer = fs.readFileSync(_path.join(examplesDirectory, name), "binary");
		callback(name, buffer);
		return endLoadFile(-1);
	}
	return JSZipUtils.getBinaryContent("../examples/" + name, function (err, data) {
		if (err) {
			throw err;
		}
		callback(name, data);
		return endLoadFile(-1);
	});
}

var startFunction = void 0;
function setStartFunction(sf) {
	allStarted = false;
	countFiles = 1;
	startFunction = sf;
}

function endLoadFile(change) {
	change = change || 0;
	countFiles += change;
	if (countFiles === 0 && allStarted === true) {
		return startFunction();
	}
}

function endsWith(str, suffix) {
	return str.indexOf(suffix, str.length - suffix.length) !== -1;
}
function startsWith(str, suffix) {
	return str.indexOf(suffix) === 0;
}
function walk(dir) {
	var results = [];
	var list = fs.readdirSync(dir);
	list.forEach(function (file) {
		file = dir + "/" + file;
		var stat = fs.statSync(file);
		if (stat && stat.isDirectory()) {
			results = results.concat(walk(file));
		} else {
			results.push(file);
		}
	});
	return results;
}

function start() {
	var fileNames = walk(examplesDirectory);
	fileNames.forEach(function (fullFileName) {
		var fileName = fullFileName.replace(examplesDirectory + "/", "");
		var callback = void 0;
		if (startsWith(fileName, ".")) {
			return;
		}
		if (endsWith(fileName, ".docx") || endsWith(fileName, ".pptx")) {
			callback = loadDocument;
		}
		if (endsWith(fileName, ".png")) {
			callback = loadImage;
		}
		if (!callback) {
			return;
		}
		loadFile(fileName, callback);
	});
	allStarted = true;
	endLoadFile(-1);
}

function setExamplesDirectory(ed) {
	examplesDirectory = ed;
}

function removeSpaces(text) {
	return text.replace(/\n|\t/g, "");
}

function makeDocx(name, content) {
	var zip = new JSZip();
	zip.file("word/document.xml", content, { createFolders: true });
	var base64 = zip.generate({ type: "string" });
	return load(name, base64, "docx", docX);
}

function createDoc(name) {
	return loadDocument(name, docX[name].loadedContent);
}

module.exports = {
	cleanError: cleanError,
	createXmlTemplaterDocx: createXmlTemplaterDocx,
	createDoc: createDoc,
	loadDocument: loadDocument,
	loadImage: loadImage,
	shouldBeSame: shouldBeSame,
	imageData: imageData,
	loadFile: loadFile,
	start: start,
	chai: chai,
	expect: expect,
	setStartFunction: setStartFunction,
	setExamplesDirectory: setExamplesDirectory,
	expectToThrow: expectToThrow,
	removeSpaces: removeSpaces,
	wrapMultiError: wrapMultiError,
	makeDocx: makeDocx
};