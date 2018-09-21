/*
Copyright 2018 The pdfcpu Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pdfcpu

import (
	"strings"
	"time"

	"github.com/hhrutter/pdfcpu/pkg/log"
)

func csvSafeString(s string) string {
	return strings.Replace(s, ";", ",", -1)
}

func handleInfoDict(ctx *Context, dict *Dict) (err error) {

	for key, value := range *dict {

		switch key {

		case "Title":
			log.Debug.Println("found Title")

		case "Author":
			log.Debug.Println("found Author")
			// Record for stats.
			ctx.Author, err = ctx.DereferenceText(value)
			if err != nil {
				return err
			}
			ctx.Author = csvSafeString(ctx.Author)

		case "Subject":
			log.Debug.Println("found Subject")

		case "Keywords":
			log.Debug.Println("found Keywords")

		case "Creator":
			log.Debug.Println("found Creator")
			// Record for stats.
			ctx.Creator, err = ctx.DereferenceText(value)
			if err != nil {
				return err
			}
			ctx.Creator = csvSafeString(ctx.Creator)

		case "Producer", "CreationDate", "ModDate":
			// pdfcpu will modify these as direct dict entries.
			log.Debug.Printf("found %s", key)
			if indRef, ok := value.(IndirectRef); ok {
				// Get rid of these extra objects.
				ctx.Optimize.DuplicateInfoObjects[int(indRef.ObjectNumber)] = true
			}

		case "Trapped":
			log.Debug.Println("found Trapped")

		default:
			log.Debug.Printf("writeInfoDict: found out of spec entry %s %v\n", key, value)

		}
	}

	return nil
}

func ensureInfoDict(ctx *Context) error {

	// => 14.3.3 Document Information Dictionary

	// Optional:
	// Title                -
	// Author               -
	// Subject              -
	// Keywords             -
	// Creator              -
	// Producer		        modified by pdfcpu
	// CreationDate	        modified by pdfcpu
	// ModDate		        modified by pdfcpu
	// Trapped              -

	now := DateString(time.Now())

	if ctx.Info == nil {

		d := NewDict()
		d.InsertString("Producer", PDFCPULongVersion)
		d.InsertString("CreationDate", now)
		d.InsertString("ModDate", now)

		indRef, err := ctx.IndRefForNewObject(d)
		if err != nil {
			return err
		}

		ctx.Info = indRef

		return nil
	}

	dict, err := ctx.DereferenceDict(*ctx.Info)
	if err != nil || dict == nil {
		return err
	}

	err = handleInfoDict(ctx, dict)
	if err != nil {
		return err
	}

	dict.Update("CreationDate", StringLiteral(now))
	dict.Update("ModDate", StringLiteral(now))
	dict.Update("Producer", StringLiteral(PDFCPULongVersion))

	return nil
}

// Write the document info object for this PDF file.
func writeDocumentInfoDict(ctx *Context) error {

	log.Debug.Printf("*** writeDocumentInfoDict begin: offset=%d ***\n", ctx.Write.Offset)

	// Note: The document info object is optional but pdfcpu ensures one.

	if ctx.Info == nil {
		log.Debug.Printf("writeDocumentInfoObject end: No info object present, offset=%d\n", ctx.Write.Offset)
		return nil
	}

	log.Debug.Printf("writeDocumentInfoObject: %s\n", *ctx.Info)

	obj := *ctx.Info

	dict, err := ctx.DereferenceDict(obj)
	if err != nil || dict == nil {
		return err
	}

	_, _, err = writeDeepObject(ctx, obj)
	if err != nil {
		return err
	}

	log.Debug.Printf("*** writeDocumentInfoDict end: offset=%d ***\n", ctx.Write.Offset)

	return nil
}
