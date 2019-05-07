var express = require('express')
var ipfsAPI = require('ipfs-api')
var multer = require('multer')
var util = require('util')
var path=require('path')
var fs = require('fs')
var bodyparser = require ('body-parser')
var mime = require("mime");
var JSZip = require('jszip');
var config = require('config');

var app = express();
var ipfs = ipfsAPI('/ip4/127.0.0.1/tcp/5001');
var router = express.Router();


var downloadPath = config.get("downloadsPath");

var log_file = fs.createWriteStream(config.get('logpath'), {flags : 'a'});
var log_stdout = process.stdout;
console.log = function(d) { //
  log_file.write(util.format(d) + '\n');
  log_stdout.write(util.format(d) + '\n');
};

process.on('uncaughtException', function (err) {
    console.error(err.stack);
    console.log(err.stack);
    console.log("Error logged and moving on");  
  });

  var filstorage =   multer.diskStorage({
    destination: function (req, file, callback) {
      console.log(req);
      callback(null, path.join(__dirname,downloadPath));
    },
    filename: function (req, file, callback) {
      //var ext = file.originalname.split('.').pop();
      callback(null,file.originalname);
      //callback(null,'sample.xlsx');
    }
  });


app.set("json spaces",0);
app.use(express.static(__dirname+downloadPath));

app.use(bodyparser.json({limit: '50mb'}));
app.use(bodyparser.urlencoded({limit: '50mb', extended: true}));
var upload = multer({ storage: filstorage});
var jsonparser = bodyparser.json();

 app.use(function(req, res, next) {
    res.header("Access-Control-Allow-Origin", "*");
    res.header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept");
    next();
  });

 /* app.use(express.static(__dirname+"/templates"));
  app.use(express.static(__dirname+"/tempOutPath")); */

  var fileupload = multer().single('uploadfile');

  app.post('/upload',jsonparser,function (req,res) {
    
      try{
        var fileHash, fileName;
        fileName = req.body.fileName;
        var _filePath =path.join(__dirname,downloadPath);

        if(fs.existsSync(path.join(__dirname,downloadPath, fileName))){
          fs.unlinkSync(path.join(__dirname,downloadPath, fileName));
        }

        fs.writeFileSync(path.join(__dirname,downloadPath, fileName),new Buffer(req.body.file));
        
        ipfs.util.addFromFs(path.join(__dirname,downloadPath,fileName),(err, result)=>{
          if (err) {
            // res.end(err.toString())       
            console.log("Upload file error to IPFS:" + err + err.stack);
            res.status(500).end(err.stack); 
          }
          if(result!=='undefined'){
            console.log("successfully uploaded the file to IPFS");
            console.log(result[0].hash);
            fileHash = result[0].hash;
            fs.unlinkSync(path.join(__dirname,downloadPath, fileName));
            //fs.unlinkSync(path.join(__dirname,downloadPath, generatedPdfName));
            res.status(200).send(fileHash);
          }
          }); 
        }
      catch (ex){
        console.log("Upload file error:" + ex);
        res.status(500).send(ex); 
      }
      
    });

    app.post('/generate',jsonparser,function (req,res) {

      if (fs.existsSync(path.join(__dirname,'tempOutPath', 'generated_doc.docx'))) {
            fs.unlinkSync(path.join(__dirname,'tempOutPath', 'generated_doc.docx'));  
        }
          
        var content = fs.readFileSync(path.join(__dirname,'templates','input.docx'), 'binary');
        var zip = new JSZip(content);
        var doc = new Docxtemplater();
        var docBookmarkContent = {};
        doc.loadZip(zip);
        console.log(req.body.length);
        console.log(req.body);

        for(var prodCount = 0; prodCount < req.body.length;prodCount++){
          var jsonObj = req.body[prodCount];
          var prodkeys = Object.keys(jsonObj);
          for (var keycount = 0;keycount<prodkeys.length; keycount++){
            docBookmarkContent[prodkeys[keycount]] =jsonObj[prodkeys[keycount]];
          }
        }
        doc.setData(docBookmarkContent);
            
        /* doc.setData({
            Customer_Name: req.body.name,
            Customer_Address:req.body.address,
            VehicleCount: req.body.vehicleCount,
            UserCount: req.body.userCount,
            MobileEyeFee: req.body.mobileEyeFee,
            MobileAppFee: req.body.mobileAppFee
        });*/
        try {
            doc.render()
        }
        catch (error) {
            var e = {
                message: error.message,
                name: error.name,
                stack: error.stack,
                properties: error.properties,
            }
            console.log(JSON.stringify({error: e}));
            res.status(500).send(error.message);
        } 
        
        var buf = doc.getZip().generate({type: 'nodebuffer'});
        fs.writeFileSync(path.join(__dirname,'tempOutPath', 'generated_doc.docx'), buf);
        var outfile = fs.createReadStream(path.join(__dirname,'tempOutPath', 'generated_doc.docx'),{encoding:'base64'});
        outfile.pipe(res);

    });


    app.get('/download', function(req,res){
      var filePath, fileHash,filename;
      fileHash = req.query.filehash;
      filename = req.query.name;
      filePath = path.join(__dirname, downloadPath,filename);

      if(fs.existsSync(filePath)){
        fs.unlinkSync(filePath);
      }
     
      ipfs.cat(fileHash,(err, stream)=>{
        if (err) {
            console.log(err.toString());
            res.status(500).send(err.toString());     
        }
        var mimetype = mime.lookup(filePath);
        var writedoc = fs.createWriteStream(filePath,{'flags':'a'});
      
        stream.on('data', function (chunk) {
          writedoc.write(chunk);
        })
      
        stream.on('error', function (err) {
          fs.unlinkSync(filePath);
          console.error('Error downloading file', err)
          res.status(500).send(err.toString());    
        })
              
          stream.on('end', function () {
          var outfile = fs.createReadStream(filePath,{encoding:'base64'});
          outfile.pipe(res);
        }) 
      })
    });

    app.post('/converttopdf',jsonparser,function (req,res) {
      
        try{
          var fileHash, fileName;
          fileName = req.body.fileName;
          var generate = req.body.generatepdf;
          var _filePath =path.join(__dirname,downloadPath);
  
          if(fs.existsSync(path.join(__dirname,downloadPath, fileName))){
            fs.unlinkSync(path.join(__dirname,downloadPath, fileName));
          }
  
          fs.writeFileSync(path.join(__dirname,downloadPath, fileName),new Buffer(req.body.file));
  
          if(fs.existsSync(path.join(__dirname,downloadPath, fileName)))
          {
            if(generate){
                //var docfileName = fileName;
                var generatedfileName = fileName.split(".")[0] + ".pdf";
                unoconv.convert(path.join(__dirname,downloadPath,fileName),'pdf', {port:2002}, function(err, filedata) {
                if(err)            
                    console.log("Error generating pdf" + err);
                else{
                  fs.writeFile(path.join(__dirname,downloadPath,generatedfileName),filedata,function(fileerror, fileresult){
                  if(fileerror){
                      console.log("Error occured in file conversion from docx to pdf.")
                  }
                  if(fileresult!=='undefined'){
                      ipfs.util.addFromFs(path.join(__dirname,downloadPath,generatedfileName),(err, result)=>{
                      if (err) {
                        console.log("Upload file error to IPFS:" + err + err.stack);
                        res.status(500).end(err.stack); 
                      }
                      if(result!=='undefined'){
                        console.log("successfully generated and uploaded the file to IPFS");
                        fileHash = result[0].hash;
                        fs.unlinkSync(path.join(__dirname,downloadPath, fileName));
                        fs.unlinkSync(path.join(__dirname,downloadPath, generatedfileName));
                        //fs.unlinkSync(path.join(__dirname,downloadPath, generatedPdfName));
                        res.status(200).send(fileHash);
                      }
                    })
                  }
                  });
                }
                });

              }
              
            }
          }
          catch (ex){
            console.log("Upload file error:" + ex);
            res.status(500).send(ex); 
          }
      });



    app.listen(9000,function () {
       // unoconv.listen();
            console.log('Node server started!')
    });