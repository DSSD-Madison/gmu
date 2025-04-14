// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.833
package components

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
)

func PDFMetadataEditForm(
	fileId string,
	originalFilename string,
	title string,
	abstract string,
	publishDate string,
	source string,
	selectedRegions []Pair,
	selectedKeywords []Pair,
	selectedAuthors []Pair,
	selectedCategories []Pair,
	csrf string,
	allRegions []db.Region,
	allKeywords []db.Keyword,
	allAuthors []db.Author,
	allCategories []db.Category,
) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Var2 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
			templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
			templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
			if !templ_7745c5c3_IsBuffer {
				defer func() {
					templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
					if templ_7745c5c3_Err == nil {
						templ_7745c5c3_Err = templ_7745c5c3_BufErr
					}
				}()
			}
			ctx = templ.InitializeContext(ctx)
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<div class=\"container max-w-2xl p-6 mx-auto mt-10 mb-10 bg-white rounded shadow-md dark:bg-gray-800\"><h1 class=\"mb-4 text-2xl font-bold text-gray-900 dark:text-white\">Edit Metadata</h1><p class=\"mb-6 text-gray-600 dark:text-gray-300\">Editing metadata for: <span class=\"font-medium text-gray-800 dark:text-white\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var3 string
			templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(originalFilename)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `web/components/metadata-edit-form.templ`, Line: 28, Col: 100}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "</span> <span class=\"block text-xs text-gray-400 break-all dark:text-gray-500\">File ID: ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var4 string
			templ_7745c5c3_Var4, templ_7745c5c3_Err = templ.JoinStringErrs(fileId)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `web/components/metadata-edit-form.templ`, Line: 29, Col: 92}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var4))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "</span></p><form hx-post=\"/save-metadata\" method=\"post\" hx-target=\"#metadata-message\" hx-swap=\"innerHTML\" hx-credentials=\"include\"><input type=\"hidden\" name=\"fileId\" value=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var5 string
			templ_7745c5c3_Var5, templ_7745c5c3_Err = templ.JoinStringErrs(fileId)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `web/components/metadata-edit-form.templ`, Line: 33, Col: 53}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var5))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 4, "\"> <input type=\"hidden\" name=\"_csrf\" value=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var6 string
			templ_7745c5c3_Var6, templ_7745c5c3_Err = templ.JoinStringErrs(csrf)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `web/components/metadata-edit-form.templ`, Line: 34, Col: 50}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var6))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 5, "\"><div class=\"mb-4\"><label for=\"title\" class=\"block mb-2 text-sm font-bold text-gray-700 dark:text-gray-200\">Title</label> <input type=\"text\" id=\"title\" name=\"title\" value=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var7 string
			templ_7745c5c3_Var7, templ_7745c5c3_Err = templ.JoinStringErrs(title)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `web/components/metadata-edit-form.templ`, Line: 38, Col: 61}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var7))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 6, "\" class=\"w-full px-3 py-2 leading-tight text-gray-700 bg-white border border-gray-300 rounded shadow appearance-none dark:border-gray-600 dark:text-gray-200 dark:bg-gray-700 focus:outline-none focus:shadow-outline\"></div><div class=\"mb-4\"><label for=\"abstract\" class=\"block mb-2 text-sm font-bold text-gray-700 dark:text-gray-200\">Abstract</label> <textarea id=\"abstract\" name=\"abstract\" rows=\"4\" class=\"w-full px-3 py-2 leading-tight text-gray-700 bg-white border border-gray-300 rounded shadow appearance-none dark:border-gray-600 dark:text-gray-200 dark:bg-gray-700 focus:outline-none focus:shadow-outline\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var8 string
			templ_7745c5c3_Var8, templ_7745c5c3_Err = templ.JoinStringErrs(abstract)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `web/components/metadata-edit-form.templ`, Line: 45, Col: 229}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var8))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 7, "</textarea></div><div class=\"grid grid-cols-1 gap-4 mb-4 md:grid-cols-2\"><div><label for=\"publish_date\" class=\"block mb-2 text-sm font-bold text-gray-700 dark:text-gray-200\">Publish Date</label> <input type=\"date\" id=\"publish_date\" name=\"publish_date\" value=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var9 string
			templ_7745c5c3_Var9, templ_7745c5c3_Err = templ.JoinStringErrs(publishDate)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `web/components/metadata-edit-form.templ`, Line: 51, Col: 82}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var9))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 8, "\" class=\"w-full px-3 py-2 leading-tight text-gray-700 bg-white border border-gray-300 rounded shadow appearance-none dark:border-gray-600 dark:text-gray-200 dark:bg-gray-700 focus:outline-none focus:shadow-outline\"></div><div><label for=\"source\" class=\"block mb-2 text-sm font-bold text-gray-700 dark:text-gray-200\">Source</label> <input type=\"text\" id=\"source\" name=\"source\" value=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var10 string
			templ_7745c5c3_Var10, templ_7745c5c3_Err = templ.JoinStringErrs(source)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `web/components/metadata-edit-form.templ`, Line: 56, Col: 65}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var10))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 9, "\" class=\"w-full px-3 py-2 leading-tight text-gray-700 bg-white border border-gray-300 rounded shadow appearance-none dark:border-gray-600 dark:text-gray-200 dark:bg-gray-700 focus:outline-none focus:shadow-outline\"><p class=\"mt-1 text-xs text-gray-500 dark:text-gray-400\">Internal reference (e.g., bucket name)</p></div></div><hr class=\"my-6 border-gray-300 dark:border-gray-600\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = TagInputJS("categories", "Category Names", "category_names", "/categories", selectedCategories).Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = TagInputJS("regions", "Region Names", "region_names", "/regions", selectedRegions).Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = TagInputJS("keywords", "Keyword Names", "keyword_names", "/keywords", selectedKeywords).Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = TagInputJS("authors", "Author Names", "author_names", "/authors", selectedAuthors).Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 10, "<div class=\"flex items-center justify-start mt-8\"><button type=\"submit\" class=\"px-4 py-2 font-bold text-white bg-blue-500 rounded hover:bg-blue-700 focus:outline-none focus:shadow-outline\">Save Metadata</button></div><div class=\"pt-2\" id=\"metadata-message\"></div></form></div><script>\n\t\t\tfunction addTag(idPrefix, fieldName, uuid, displayName) {\n\t\t\t\tconst tagValue = uuid.trim();\n\t\t\t\tconst tagLabel = displayName.trim();\n\t\t\t\tif (!tagValue || !tagLabel) return;\n\n\t\t\t\tconst container = document.getElementById(`${idPrefix}-tags-display`)?.closest('.tag-input-container');\n\t\t\t\tif (!container) return;\n\n\t\t\t\tconst tagsDisplay = container.querySelector(`#${idPrefix}-tags-display`);\n\t\t\t\tconst hiddenInputsContainer = container.querySelector(`#${idPrefix}-hidden-inputs`);\n\t\t\t\tconst searchInput = container.querySelector(`#${idPrefix}-search-input`);\n\t\t\t\tconst suggestionsContainer = container.querySelector(`#${idPrefix}-suggestions`);\n\n\t\t\t\tconst existingInput = hiddenInputsContainer.querySelector(`input[name=\"${fieldName}\"][value=\"${CSS.escape(tagValue)}\"]`);\n\t\t\t\tif (existingInput) return;\n\n\t\t\t\tconst hiddenInput = document.createElement('input');\n\t\t\t\thiddenInput.type = 'hidden';\n\t\t\t\thiddenInput.name = fieldName;\n\t\t\t\thiddenInput.value = tagValue;\n\t\t\t\thiddenInput.setAttribute('data-tag-value', tagValue);\n\t\t\t\thiddenInputsContainer.appendChild(hiddenInput);\n\n\t\t\t\tconst tagSpan = document.createElement('span');\n\t\t\t\ttagSpan.setAttribute('data-tag-value', tagValue);\n\t\t\t\ttagSpan.setAttribute('data-id-prefix', idPrefix);\n\t\t\t\ttagSpan.setAttribute('data-field-name', fieldName);\n\t\t\t\ttagSpan.className = 'tag-item bg-blue-100 text-blue-800 text-xs font-medium me-2 px-2.5 py-0.5 rounded dark:bg-blue-900 dark:text-blue-300 inline-flex items-center';\n\t\t\t\ttagSpan.textContent = tagLabel + ' ';\n\n\t\t\t\tconst removeButton = document.createElement('button');\n\t\t\t\tremoveButton.type = 'button';\n\t\t\t\tremoveButton.className = 'ml-1 text-blue-600 hover:text-blue-400 focus:outline-none';\n\t\t\t\tremoveButton.innerHTML = '×';\n\t\t\t\tremoveButton.setAttribute('aria-label', `Remove ${tagLabel}`);\n\t\t\t\tremoveButton.onclick = function () { removeTag(this); };\n\t\t\t\ttagSpan.appendChild(removeButton);\n\n\t\t\t\tconst placeholder = tagsDisplay.querySelector('.tag-placeholder');\n\t\t\t\tif (placeholder) placeholder.remove();\n\t\t\t\ttagsDisplay.appendChild(tagSpan);\n\n\t\t\t\tsearchInput.value = '';\n\t\t\t\tsuggestionsContainer.innerHTML = '';\n\t\t\t\tsearchInput.focus();\n\t\t\t}\n\n\t\t\tfunction removeTag(buttonElement) {\n\t\t\t\tconst tagSpan = buttonElement.closest('.tag-item');\n\t\t\t\tif (!tagSpan) return;\n\n\t\t\t\tconst tagValue = tagSpan.getAttribute('data-tag-value');\n\t\t\t\tconst idPrefix = tagSpan.getAttribute('data-id-prefix');\n\t\t\t\tconst fieldName = tagSpan.getAttribute('data-field-name');\n\n\t\t\t\tconst container = tagSpan.closest('.tag-input-container');\n\t\t\t\tif (!container || !tagValue || !idPrefix || !fieldName) return;\n\n\t\t\t\tconst hiddenInputsContainer = container.querySelector(`#${idPrefix}-hidden-inputs`);\n\t\t\t\tconst tagsDisplay = container.querySelector(`#${idPrefix}-tags-display`);\n\n\t\t\t\tconst hiddenInput = hiddenInputsContainer?.querySelector(`input[name=\"${fieldName}\"][data-tag-value=\"${CSS.escape(tagValue)}\"]`);\n\t\t\t\tif (hiddenInput) hiddenInput.remove();\n\n\t\t\t\ttagSpan.remove();\n\n\t\t\t\tif (tagsDisplay && !tagsDisplay.querySelector('.tag-item')) {\n\t\t\t\t\tconst placeholder = document.createElement('span');\n\t\t\t\t\tplaceholder.className = 'tag-placeholder text-xs text-gray-400 italic p-1';\n\t\t\t\t\tlet labelText = 'items';\n\t\t\t\t\tconst labelElement = container.querySelector(`label[for='${idPrefix}-search-input']`);\n\t\t\t\t\tif (labelElement) {\n\t\t\t\t\t\tlabelText = labelElement.textContent.replace(/\\s+Names$/i, '').toLowerCase();\n\t\t\t\t\t}\n\t\t\t\t\tplaceholder.textContent = `No ${labelText} added yet.`;\n\t\t\t\t\ttagsDisplay.appendChild(placeholder);\n\t\t\t\t}\n\t\t\t}\n\n\t\t\tdocument.addEventListener('click', function(event) {\n\t\t\t\tconst allTagContainers = document.querySelectorAll('.tag-input-container');\n\t\t\t\tallTagContainers.forEach(container => {\n\t\t\t\t\tconst suggestionsDivId = container.querySelector('input[type=text]').id.replace('-search-input', '-suggestions');\n\t\t\t\t\tconst suggestionsDiv = container.querySelector(`#${suggestionsDivId}`);\n\t\t\t\t\tif (suggestionsDiv && !container.contains(event.target)) {\n\t\t\t\t\t\tsuggestionsDiv.innerHTML = '';\n\t\t\t\t\t}\n\t\t\t\t});\n\t\t\t});\n\n\t\t\tdocument.addEventListener(\"input\", function (event) {\n\t\t\t\tconst input = event.target;\n\t\t\t\tif (!input.matches(\".tag-search-input\")) return;\n\n\t\t\t\tconst container = input.closest(\".tag-input-container\");\n\t\t\t\tif (!container) return;\n\n\t\t\t\tconst suggestionsContainer = container.querySelector(`#${input.id.replace(\"-search-input\", \"-suggestions\")}`);\n\t\t\t\tconst query = input.value.trim();\n\t\t\t\tif (!query) {\n\t\t\t\t\tsuggestionsContainer.innerHTML = '';\n\t\t\t\t\treturn;\n\t\t\t\t}\n\n\t\t\t\t// Load suggestions from the server\n\t\t\t\tconst endpoint = input.getAttribute(\"data-endpoint\");\n\t\t\t\tfetch(`${endpoint}?q=${encodeURIComponent(query)}`)\n\t\t\t\t\t.then(res => res.json())\n\t\t\t\t\t.then(data => {\n\t\t\t\t\t\tsuggestionsContainer.innerHTML = \"\";\n\n\t\t\t\t\t\tconst alreadyExists = data.some(item => item.name.toLowerCase() === query.toLowerCase());\n\t\t\t\t\t\tif (!alreadyExists) {\n\t\t\t\t\t\t\tconst newOption = document.createElement(\"div\");\n\t\t\t\t\t\t\tnewOption.className = \"tag-suggestion text-green-700 hover:bg-green-100 dark:hover:bg-gray-700 px-2 py-1 cursor-pointer font-semibold\";\n\t\t\t\t\t\t\tnewOption.textContent = `+ Create \"${query}\"`;\n\t\t\t\t\t\t\tnewOption.onclick = function () {\n\t\t\t\t\t\t\t\taddTag(\n\t\t\t\t\t\t\t\t\tinput.getAttribute(\"data-id-prefix\"),\n\t\t\t\t\t\t\t\t\tinput.getAttribute(\"data-field-name\"),\n\t\t\t\t\t\t\t\t\t`new:${query}`,\n\t\t\t\t\t\t\t\t\tquery\n\t\t\t\t\t\t\t\t);\n\t\t\t\t\t\t\t};\n\t\t\t\t\t\t\tsuggestionsContainer.appendChild(newOption);\n\t\t\t\t\t\t}\n\n\t\t\t\t\t\tfor (const item of data) {\n\t\t\t\t\t\t\tconst div = document.createElement(\"div\");\n\t\t\t\t\t\t\tdiv.className = \"tag-suggestion px-2 py-1 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700\";\n\t\t\t\t\t\t\tdiv.textContent = item.name;\n\t\t\t\t\t\t\tdiv.onclick = function () {\n\t\t\t\t\t\t\t\taddTag(\n\t\t\t\t\t\t\t\t\tinput.getAttribute(\"data-id-prefix\"),\n\t\t\t\t\t\t\t\t\tinput.getAttribute(\"data-field-name\"),\n\t\t\t\t\t\t\t\t\titem.id,\n\t\t\t\t\t\t\t\t\titem.name\n\t\t\t\t\t\t\t\t);\n\t\t\t\t\t\t\t};\n\t\t\t\t\t\t\tsuggestionsContainer.appendChild(div);\n\t\t\t\t\t\t}\n\t\t\t\t\t});\n\t\t\t});\n\t\t</script>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			return nil
		})
		templ_7745c5c3_Err = Base("Edit PDF Metadata").Render(templ.WithChildren(ctx, templ_7745c5c3_Var2), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
