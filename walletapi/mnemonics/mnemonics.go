// Copyright 2017-2018 DERO Project. All rights reserved.
// Use of this source code in any form is governed by RESEARCH license.
// license can be found in the LICENSE file.
// GPG: 0F39 E425 8C65 3947 702A  8234 08B2 0360 A03A 9DE8
//
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY
// EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL
// THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
// PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
// STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF
// THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package mnemonics

import "fmt"
import "strings"
import "encoding/binary"
import "unicode/utf8"

//import "github.com/romana/rlog"

import "hash/crc32"

import "github.com/deroproject/derosuite/crypto"

type Language struct {
	Name                 string   // Name of the language
	Name_English         string   // Name of the language in english
	Unique_Prefix_Length int      // number of utf8 chars (not bytes) to use for checksum
	Words                []string // 1626 words
}

// any language needs to be added to this array
var Languages = []Language{
	Mnemonics_English,
	Mnemonics_Japanese,
	Mnemonics_Chinese_Simplified,
	Mnemonics_Dutch,
	Mnemonics_Esperanto,
	Mnemonics_Russian,
	Mnemonics_Spanish,
	Mnemonics_Portuguese,
	Mnemonics_French,
	Mnemonics_German,
	Mnemonics_Italian,
}

const SEED_LENGTH = 24 // checksum seeds are 24 + 1 = 25 words long

// while init check whether all languages have  necessary data

func init() {
	for i := range Languages {
		if len(Languages[i].Words) != 1626 {
			panic(fmt.Sprintf("%s language only has %d words, should have 1626", Languages[i].Name_English, len(Languages[i].Words)))
		}
	}
}

// return all the languages support by us
func Language_List() (list []string) {
	for i := range Languages {
		list = append(list, Languages[i].Name)
	}
	return
}

//this function converts a list of words to a key
func Words_To_Key(words_line string) (language_name string, key crypto.Key, err error) {

	checksum_present := false
	words := strings.Fields(words_line)
	//rlog.Tracef(1, "len of words %d", words)

	// if seed size is not 24 or 25, return err
	if len(words) != SEED_LENGTH && len(words) != (SEED_LENGTH+1) {
		err = fmt.Errorf("Invalid Seed")
		return
	}

	// if checksum is present consider it so
	if len(words) == (SEED_LENGTH + 1) {
		checksum_present = true
	}

	indices, language_index, wordlist_length, found := Find_indices(words)

	if !found {
		return language_name, key, fmt.Errorf("Seed not found in any Language")
	}

	language_name = Languages[language_index].Name

	if checksum_present { // we need language unique prefix to validate checksum
		if !Verify_Checksum(words, Languages[language_index].Unique_Prefix_Length) {
			return language_name, key, fmt.Errorf("Seed Checksum failed")
		}
	}

	// key = make([]byte,(SEED_LENGTH/3)*4,(SEED_LENGTH/3)*4) // our keys are  32 bytes

	// map 3 words to 4 bytes each
	// so 24 words = 32 bytes
	for i := 0; i < (SEED_LENGTH / 3); i++ {
		w1 := indices[i*3]
		w2 := indices[i*3+1]
		w3 := indices[i*3+2]

		val := w1 + wordlist_length*(((wordlist_length-w1)+w2)%wordlist_length) +
			wordlist_length*wordlist_length*(((wordlist_length-w2)+w3)%wordlist_length)

			// sanity check, this can never occur
		if (val % wordlist_length) != w1 {
			panic("Word list error")
		}

		value_32bit := uint32(val)

		binary.LittleEndian.PutUint32(key[i*4:], value_32bit) // place key into output container

		//memcpy(dst.data + i * 4, &val, 4);  // copy 4 bytes to position
	}
	//fmt.Printf("words %+v\n", indices)
	//fmt.Printf("key %x\n", key)

	return
}

// this will map the key to recovery words from the spcific language
// language must exist,if not we return english
func Key_To_Words(key crypto.Key, language string) (words_line string) {
	var words []string // all words are appended here

	l_index := 0
	for i := range Languages {
		if Languages[i].Name == language {
			l_index = i
			break
		}
	}

	// total numbers of words in specified language dictionary
	word_list_length := uint32(len(Languages[l_index].Words))

	// 8 bytes -> 3 words.  8 digits base 16 -> 3 digits base 1626
	// for (unsigned int i=0; i < sizeof(src.data)/4; i++, words += ' ')
	for i := 0; i < (len(key) / 4); i++ {

		val := binary.LittleEndian.Uint32(key[i*4:])

		w1 := val % word_list_length
		w2 := ((val / word_list_length) + w1) % word_list_length
		w3 := (((val / word_list_length) / word_list_length) + w2) % word_list_length

		words = append(words, Languages[l_index].Words[w1])
		words = append(words, Languages[l_index].Words[w2])
		words = append(words, Languages[l_index].Words[w3])
	}

	checksum_index, err := Calculate_Checksum_Index(words, Languages[l_index].Unique_Prefix_Length)
	if err != nil {
		//fmt.Printf("Checksum index failed")
		return

	} else {
		// append checksum word
		words = append(words, words[checksum_index])

	}

	words_line = strings.Join(words, " ")

	//fmt.Printf("words %s \n", words_line)

	return

}

// find language and indices
// all words should be from the same languages ( words do not cross language boundary )
// indices = position where word was found
// language = which language the seed is in
// word_list_count = total words in the specified language

func Find_indices(words []string) (indices []uint64, language_index int, word_list_count uint64, found bool) {

	for i := range Languages {
		var local_indices []uint64 // we build a local copy

		// create a map from words , for finding the words faster
		language_map := map[string]int{}
		for j := 0; j < len(Languages[i].Words); j++ {
			language_map[Languages[i].Words[j]] = j
		}

		// now lets loop through all the user supplied words
		for j := 0; j < len(words); j++ {
			if v, ok := language_map[words[j]]; ok {
				local_indices = append(local_indices, uint64(v))
			} else { // word has missed, this cannot be our language
				goto try_another_language
			}
		}

		// if we have found all the words, this is our language of seed words
		// stop processing and return all data
		return local_indices, i, uint64(len(Languages[i].Words)), true

	try_another_language:
	}

	// we are here, means we could locate any language which contains all the seed words
	// return empty
	return
}

// calculate a checksum on first 24 words
// checksum is calculated as follows
// take prefix_len chars ( not bytes) from first 24 words and concatenate them
// calculate crc of resultant concatenated bytes
// take mod of SEED_LENGTH 24, to get the checksum word

func Calculate_Checksum_Index(words []string, prefix_len int) (uint32, error) {
	var trimmed_runes []rune

	if len(words) != SEED_LENGTH {
		return 0, fmt.Errorf("Words not equal to seed length")
	}

	for i := range words {
		if utf8.RuneCountInString(words[i]) > prefix_len { // take first prefix_len utf8 chars
			trimmed_runes = append(trimmed_runes, ([]rune(words[i]))[0:prefix_len]...)
		} else {
			trimmed_runes = append(trimmed_runes, ([]rune(words[i]))...) /// add entire string
		}

	}

	checksum := crc32.ChecksumIEEE([]byte(string(trimmed_runes)))

	//fmt.Printf("trimmed words %s  %d \n", string(trimmed_runes), checksum)

	return checksum % SEED_LENGTH, nil

}

// for verification, we need all 25 words
// calculate checksum and verify whether match
func Verify_Checksum(words []string, prefix_len int) bool {

	if len(words) != (SEED_LENGTH + 1) {
		return false // Checksum word is not present, we cannot verify
	}

	checksum_index, err := Calculate_Checksum_Index(words[:len(words)-1], prefix_len)
	if err != nil {
		return false
	}
	calculated_checksum_word := words[checksum_index]
	checksum_word := words[SEED_LENGTH]

	if calculated_checksum_word == checksum_word {
		return true
	}

	return false
}
